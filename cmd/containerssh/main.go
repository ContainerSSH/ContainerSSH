package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/containerssh/configuration/v3"
	"github.com/containerssh/health"
	"github.com/containerssh/log"
	"github.com/containerssh/service"
	"github.com/containerssh/structutils"

	"github.com/containerssh/containerssh"
)

func main() {
	config := configuration.AppConfig{}
	structutils.Defaults(&config)

	loggerFactory := log.NewLoggerFactory()
	logger, err := loggerFactory.Make(
		config.Log,
	)
	if err != nil {
		panic(err)
	}

	logger = logger.WithLabel("module", "core")

	configFile, actionDumpConfig, actionLicenses, actionHealthCheck := getArguments()

	if configFile == "" {
		configFile = "config.yaml"
	}
	realConfigFile, err := filepath.Abs(configFile)
	if err != nil {
		logger.Critical(log.Wrap(
			err,
			containerssh.EConfig,
			"Failed to fetch absolute path for configuration file %s",
			configFile,
		))
		os.Exit(1)
	}
	configFile = realConfigFile
	if err = readConfigFile(configFile, loggerFactory, &config); err != nil {
		logger.Critical(log.Wrap(
			err,
			containerssh.EConfig,
			"Invalid configuration in file %s",
			configFile,
		))
		os.Exit(1)
	}

	configuredLogger, err := loggerFactory.Make(
		config.Log,
	)
	if err != nil {
		logger.Critical(err)
		os.Exit(1)
	}
	configuredLogger.Debug(log.NewMessage(containerssh.MConfigFile, "Using configuration file %s...", configFile))

	switch {
	case actionDumpConfig:
		runDumpConfig(config, configuredLogger)
	case actionLicenses:
		runActionLicenses(configuredLogger)
	case actionHealthCheck:
		runHealthCheck(config, configuredLogger)
	default:
		runContainerSSH(loggerFactory, configuredLogger, config, configFile)
	}
}

func runHealthCheck(config configuration.AppConfig, logger log.Logger) {
	if err := healthCheck(config, logger); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}
	logger.Info(log.NewMessage(containerssh.MHealthCheckSuccessful, "Health check successful."))
	os.Exit(0)
}

func runActionLicenses(logger log.Logger) {
	if err := printLicenses(os.Stdout); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runDumpConfig(config configuration.AppConfig, logger log.Logger) {
	if err := dumpConfig(os.Stdout, logger, &config); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runContainerSSH(
	loggerFactory log.LoggerFactory,
	logger log.Logger,
	config configuration.AppConfig,
	configFile string,
) {
	if len(config.SSH.HostKeys) == 0 {
		logger.Warning(
			log.NewMessage(
				containerssh.ECoreNoHostKeys,
				"No host keys found in configuration, generating temporary host keys and updating configuration...",
			),
		)
		if err := generateHostKeys(configFile, &config, logger); err != nil {
			logger.Critical(log.Wrap(
				err,
				containerssh.ECoreHostKeyGenerationFailed,
				"failed to generate host keys",
			))
			os.Exit(1)
		}
	}

	if err := startServices(config, loggerFactory); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func getArguments() (string, bool, bool, bool) {
	configFile := ""
	actionDumpConfig := false
	actionLicenses := false
	healthCheck := false
	flag.StringVar(
		&configFile,
		"config",
		"",
		"Configuration file to load (has to end in .yaml, .yml, or .json)",
	)
	flag.BoolVar(
		&actionDumpConfig,
		"dump-config",
		false,
		"Dump configuration and exit",
	)
	flag.BoolVar(
		&actionLicenses,
		"licenses",
		false,
		"Print license information",
	)
	flag.BoolVar(
		&healthCheck,
		"healthcheck",
		false,
		"Run health check",
	)
	flag.Parse()
	return configFile, actionDumpConfig, actionLicenses, healthCheck
}

func startServices(config configuration.AppConfig, loggerFactory log.LoggerFactory) error {
	pool, lifecycle, err := containerssh.New(config, loggerFactory)
	if err != nil {
		return err
	}

	return startPool(pool, lifecycle)
}

func startPool(pool containerssh.Service, lifecycle service.Lifecycle) error {
	starting := make(chan struct{})
	lifecycle.OnStarting(
		func(s service.Service, l service.Lifecycle) {
			starting <- struct{}{}
		},
	)
	go func() {
		_ = lifecycle.Run()
	}()

	<-starting

	exitSignalList := []os.Signal{os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM}
	rotateSignalList := []os.Signal{syscall.SIGHUP}
	exitSignals := make(chan os.Signal, 1)
	rotateSignals := make(chan os.Signal, 1)
	signal.Notify(exitSignals, exitSignalList...)
	signal.Notify(rotateSignals, rotateSignalList...)
	go func() {
		if _, ok := <-exitSignals; ok {
			// ok means the channel wasn't closed
			shutdownContext, cancelFunc := context.WithTimeout(
				context.Background(),
				20*time.Second,
			)
			defer cancelFunc()
			lifecycle.Stop(
				shutdownContext,
			)
		}
	}()
	go func() {
		for {
			if _, ok := <-rotateSignals; ok {
				err := pool.RotateLogs()
				if err != nil {
					panic(err)
				}
			} else {
				break
			}
		}
	}()
	err := lifecycle.Wait()
	signal.Ignore(rotateSignalList...)
	signal.Ignore(exitSignalList...)
	close(exitSignals)
	return err
}

func generateHostKeys(configFile string, config *configuration.AppConfig, logger log.Logger) error {
	if err := config.SSH.GenerateHostKey(); err != nil {
		return err
	}

	tmpFile := fmt.Sprintf("%s~", configFile)
	fh, err := os.Create(tmpFile)
	if err != nil {
		logger.Warning(log.Wrap(
			err,
			containerssh.ECannotWriteConfigFile,
			"Cannot create temporary configuration file at %s with updated host keys.",
			tmpFile,
		).Label("tmpFile", configFile))
		return nil
	}
	format := getConfigFileFormat(configFile)
	saver, err := configuration.NewWriterSaver(fh, logger, format)
	if err != nil {
		_ = fh.Close()
		logger.Warning(log.Wrap(
			err,
			containerssh.ECannotWriteConfigFile,
			"Cannot initialize temporary configuration file at %s with updated host keys.",
			tmpFile,
		).Label("tmpFile", configFile))
		return nil
	}
	if err := saver.Save(config); err != nil {
		_ = fh.Close()
		logger.Warning(log.Wrap(
			err,
			containerssh.ECannotWriteConfigFile,
			"Cannot save temporary configuration file at %s with updated host keys.",
			tmpFile,
		).Label("tmpFile", configFile))
		return nil
	}
	if err := fh.Close(); err != nil {
		logger.Warning(log.Wrap(err, containerssh.ECannotWriteConfigFile, "Cannot close temporary configuration file at %s with updated host keys.", tmpFile).Label("tmpFile", configFile))
		return nil
	}

	if err := os.Rename(tmpFile, configFile); err != nil {
		logger.Warning(log.Wrap(err, containerssh.ECannotWriteConfigFile, "Failed to rename temporary file %s to %s with updated host keys.", tmpFile, configFile).Label("file", configFile).Label("tmpFile", tmpFile))
		return fmt.Errorf("failed to rename temporary file %s to %s (%w)", tmpFile, configFile, err)
	}

	return nil
}

func healthCheck(config configuration.AppConfig, logger log.Logger) error {
	healthClient, err := health.NewClient(config.Health, logger)
	if err != nil {
		return err
	}
	if healthClient == nil {
		return nil
	}
	if !healthClient.Run() {
		return log.NewMessage(containerssh.EHealthCheckFailed, "Health check failed")
	}
	return nil
}

func printLicenses(writer io.Writer) error {
	var buffer bytes.Buffer

	buffer.WriteString("# The ContainerSSH license\n\n")
	licenseData, err := ioutil.ReadFile("LICENSE.md")
	if err != nil {
		return fmt.Errorf("failed to read LICENSE.md (%w)", err)
	}
	buffer.Write(licenseData)
	buffer.WriteString("\n")
	noticeData, err := ioutil.ReadFile("NOTICE.md")
	if err != nil {
		return fmt.Errorf("failed to read NOTICE.md (%w)", err)
	}
	buffer.Write(noticeData)
	buffer.WriteString("\n")
	if _, err := writer.Write(buffer.Bytes()); err != nil {
		return fmt.Errorf("failed to write licenes information (%w)", err)
	}
	return nil
}

func dumpConfig(writer io.Writer, logger log.Logger, config *configuration.AppConfig) error {
	saver, err := configuration.NewWriterSaver(writer, logger, configuration.FormatYAML)
	if err != nil {
		return err
	}
	if err := saver.Save(config); err != nil {
		return err
	}
	return nil
}

func readConfigFile(
	configFile string,
	loggerFactory log.LoggerFactory,
	config *configuration.AppConfig,
) error {
	configLogger, err := loggerFactory.Make(
		config.Log,
	)
	if err != nil {
		return err
	}
	configFH, err := os.Open(configFile)
	if err != nil {
		return err
	}
	format := getConfigFileFormat(configFile)
	configLoader, err := configuration.NewReaderLoader(configFH, configLogger, format)
	if err != nil {
		return err
	}
	if err := configLoader.Load(context.Background(), config); err != nil {
		return fmt.Errorf("failed to read configuration file %s (%w)", configFile, err)
	}
	return nil
}

func getConfigFileFormat(configFile string) configuration.Format {
	var format configuration.Format
	if strings.HasSuffix(configFile, ".json") {
		format = configuration.FormatJSON
	} else {
		format = configuration.FormatYAML
	}
	return format
}
