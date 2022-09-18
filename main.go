package libcontainerssh

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"go.containerssh.io/libcontainerssh/config"
	internalConfig "go.containerssh.io/libcontainerssh/internal/config"
	"go.containerssh.io/libcontainerssh/internal/health"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/service"
)

// Main is a helper function to start a standard ContainerSSH instance. It should be used as the outer-most function
// and should never be used as an embedding technique.
func Main() {
	cfg := config.AppConfig{}
	cfg.Default()

	loggerFactory := log.NewLoggerFactory()
	logger, err := loggerFactory.Make(
		cfg.Log,
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
		logger.Critical(
			message.Wrap(
				err,
				message.ECoreConfig,
				"Failed to fetch absolute path for configuration file %s",
				configFile,
			))
		os.Exit(1)
	}
	configFile = realConfigFile
	if err = readConfigFile(configFile, loggerFactory, &cfg); err != nil {
		logger.Critical(
			message.Wrap(
				err,
				message.ECoreConfig,
				"Invalid configuration in file %s",
				configFile,
			))
		os.Exit(1)
	}

	configuredLogger, err := loggerFactory.Make(
		cfg.Log,
	)
	if err != nil {
		logger.Critical(err)
		os.Exit(1)
	}
	configuredLogger.Debug(message.NewMessage(message.MCoreConfigFile, "Using configuration file %s...", configFile))

	switch {
	case actionDumpConfig:
		runDumpConfig(cfg, configuredLogger)
	case actionLicenses:
		runActionLicenses(configuredLogger)
	case actionHealthCheck:
		runHealthCheck(cfg, configuredLogger)
	default:
		runContainerSSH(loggerFactory, configuredLogger, cfg, configFile)
	}
}

func runHealthCheck(cfg config.AppConfig, logger log.Logger) {
	if err := healthCheck(cfg, logger); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}
	logger.Info(message.NewMessage(message.MCoreHealthCheckSuccessful, "Health check successful."))
	os.Exit(0)
}

func runActionLicenses(logger log.Logger) {
	if err := printLicenses(os.Stdout); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runDumpConfig(cfg config.AppConfig, logger log.Logger) {
	if err := dumpConfig(os.Stdout, logger, &cfg); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runContainerSSH(
	loggerFactory log.LoggerFactory,
	logger log.Logger,
	cfg config.AppConfig,
	configFile string,
) {
	if len(cfg.SSH.HostKeys) == 0 {
		logger.Warning(
			message.NewMessage(
				message.ECoreNoHostKeys,
				"No host keys found in configuration, generating temporary host keys and updating configuration...",
			),
		)
		if err := generateHostKeys(configFile, &cfg, logger); err != nil {
			logger.Critical(
				message.Wrap(
					err,
					message.ECoreHostKeyGenerationFailed,
					"failed to generate host keys",
				))
			os.Exit(1)
		}
	}

	if err := startServices(cfg, loggerFactory); err != nil {
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

func startServices(cfg config.AppConfig, loggerFactory log.LoggerFactory) error {
	pool, lifecycle, err := New(cfg, loggerFactory)
	if err != nil {
		return err
	}

	return startPool(pool, lifecycle)
}

func startPool(pool Service, lifecycle service.Lifecycle) error {
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

func generateHostKeys(configFile string, cfg *config.AppConfig, logger log.Logger) error {
	if err := cfg.SSH.GenerateHostKey(); err != nil {
		return err
	}

	tmpFile := fmt.Sprintf("%s~", configFile)
	fh, err := os.Create(tmpFile)
	if err != nil {
		logger.Warning(
			message.Wrap(
				err,
				message.ECannotWriteConfigFile,
				"Cannot create temporary configuration file at %s with updated host keys.",
				tmpFile,
			).Label("tmpFile", configFile))
		return nil
	}
	format := getConfigFileFormat(configFile)
	saver, err := internalConfig.NewWriterSaver(fh, logger, format)
	if err != nil {
		_ = fh.Close()
		logger.Warning(
			message.Wrap(
				err,
				message.ECannotWriteConfigFile,
				"Cannot initialize temporary configuration file at %s with updated host keys.",
				tmpFile,
			).Label("tmpFile", configFile))
		return nil
	}
	if err := saver.Save(cfg); err != nil {
		_ = fh.Close()
		logger.Warning(
			message.Wrap(
				err,
				message.ECannotWriteConfigFile,
				"Cannot save temporary configuration file at %s with updated host keys.",
				tmpFile,
			).Label("tmpFile", configFile))
		return nil
	}
	if err := fh.Close(); err != nil {
		logger.Warning(message.Wrap(err,
			message.ECannotWriteConfigFile, "Cannot close temporary configuration file at %s with updated host keys.", tmpFile).Label("tmpFile", configFile))
		return nil
	}

	if err := os.Rename(tmpFile, configFile); err != nil {
		logger.Warning(message.Wrap(err,
			message.ECannotWriteConfigFile, "Failed to rename temporary file %s to %s with updated host keys.", tmpFile, configFile).Label("file", configFile).Label("tmpFile", tmpFile))
		return fmt.Errorf("failed to rename temporary file %s to %s (%w)", tmpFile, configFile, err)
	}

	return nil
}

func healthCheck(cfg config.AppConfig, logger log.Logger) error {
	healthClient, err := health.NewClient(cfg.Health, logger)
	if err != nil {
		return err
	}
	if healthClient == nil {
		return nil
	}
	if !healthClient.Run() {
		return message.NewMessage(message.ECoreHealthCheckFailed, "Health check failed")
	}
	return nil
}

func printLicenses(writer io.Writer) error {
	var buffer bytes.Buffer

	buffer.WriteString("# The ContainerSSH license\n\n")
	licenseData, err := os.ReadFile("LICENSE.md")
	if err != nil {
		return fmt.Errorf("failed to read LICENSE.md (%w)", err)
	}
	buffer.Write(licenseData)
	buffer.WriteString("\n")
	noticeData, err := os.ReadFile("NOTICE.md")
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

func dumpConfig(writer io.Writer, logger log.Logger, cfg *config.AppConfig) error {
	saver, err := internalConfig.NewWriterSaver(writer, logger, internalConfig.FormatYAML)
	if err != nil {
		return err
	}
	if err := saver.Save(cfg); err != nil {
		return err
	}
	return nil
}

func readConfigFile(
	configFile string,
	loggerFactory log.LoggerFactory,
	cfg *config.AppConfig,
) error {
	configLogger, err := loggerFactory.Make(
		cfg.Log,
	)
	if err != nil {
		return err
	}
	// File inclusion is desired here, no gosec issue.
	configFH, err := os.Open(configFile) //nolint:gosec
	if err != nil {
		return err
	}
	format := getConfigFileFormat(configFile)
	configLoader, err := internalConfig.NewReaderLoader(configFH, configLogger, format)
	if err != nil {
		return err
	}
	if err := configLoader.Load(context.Background(), cfg); err != nil {
		return fmt.Errorf("failed to read configuration file %s (%w)", configFile, err)
	}
	return nil
}

func getConfigFileFormat(configFile string) internalConfig.Format {
	var format internalConfig.Format
	if strings.HasSuffix(configFile, ".json") {
		format = internalConfig.FormatJSON
	} else {
		format = internalConfig.FormatYAML
	}
	return format
}
