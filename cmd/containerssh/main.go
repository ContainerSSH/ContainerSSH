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

	"github.com/containerssh/configuration"
	"github.com/containerssh/log"
	"github.com/containerssh/service"
	"github.com/containerssh/structutils"

	"github.com/containerssh/containerssh"
)

type sureFireWriter struct {
	backend io.Writer
}

func (s *sureFireWriter) Write(p []byte) (n int, err error) {
	n, err = s.backend.Write(p)
	if err != nil {
		// Ignore errors
		return len(p), nil
	}
	return n, nil
}

func main() {
	config := configuration.AppConfig{}
	structutils.Defaults(&config)

	loggerFactory := log.NewFactory(&sureFireWriter{os.Stdout})
	logger, err := loggerFactory.Make(
		config.Log,
		"",
	)
	if err != nil {
		panic(err)
	}

	configFile, actionDumpConfig, actionLicenses := getArguments()

	if configFile != "" {
		if err = readConfigFile(configFile, loggerFactory, &config); err != nil {
			logger.Criticalf("failed to read configuration file %s (%v)", configFile, err)
			os.Exit(1)
		}
	} else {
		configFile, err = filepath.Abs("./config.yaml")
		if err != nil {
			logger.Criticalf("failed to fetch current directory")
			os.Exit(1)
		}
	}

	if actionDumpConfig {
		if err = dumpConfig(os.Stdout, loggerFactory, &config); err != nil {
			logger.Criticalf("error dumping config (%v)", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if actionLicenses {
		if err := printLicenses(os.Stdout); err != nil {
			logger.Criticale(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(config.SSH.HostKeys) == 0 {

		logger.Noticef("No host keys found in configuration, generating host keys...")
		if err := generateHostKeys(configFile, &config, logger); err != nil {
			logger.Criticalf("failed to generate host keys (%v)", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := startServices(config, logger, loggerFactory); err != nil {
		logger.Criticale(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func getArguments() (string, bool, bool) {
	configFile := ""
	actionDumpConfig := false
	actionLicenses := false
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
	flag.Parse()
	return configFile, actionDumpConfig, actionLicenses
}

func startServices(config configuration.AppConfig, logger log.Logger, loggerFactory log.LoggerFactory) error {
	pool, err := containerssh.New(config, loggerFactory)
	if err != nil {
		return err
	}

	return startPool(pool, logger)
}

func startPool(pool service.Service, logger log.Logger) error {
	lifecycle := service.NewLifecycle(pool)
	starting := make(chan struct{})
	lifecycle.OnStarting(
		func(s service.Service, l service.Lifecycle) {
			logger.Noticef("Services are starting...")
			starting <- struct{}{}
		},
	)
	lifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			logger.Noticef("All services are now running.")
		},
	)
	lifecycle.OnStopping(
		func(s service.Service, l service.Lifecycle, shutdownContext context.Context) {
			logger.Noticef("Services are shutting down...")
		},
	)

	go func() {
		_ = lifecycle.Run()
	}()

	<-starting

	signals := make(chan os.Signal, 1)
	signalList := []os.Signal{os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP}
	signal.Notify(signals, signalList...)
	go func() {
		if _, ok := <-signals; ok {
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
	err := lifecycle.Wait()
	signal.Ignore(signalList...)
	close(signals)
	return err
}

func generateHostKeys(configFile string, config *configuration.AppConfig, logger log.Logger) error {
	if err := config.SSH.GenerateHostKey(); err != nil {
		return err
	}

	dir := filepath.Base(configFile)
	hostKeyFile := filepath.Join(dir, "ssh_host_rsa_key")

	tmpFile := fmt.Sprintf("%s~", hostKeyFile)
	fh, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to open temporary file %s for writing (%w)", tmpFile, err)
	}
	format := getConfigFileFormat(hostKeyFile)
	saver, err := configuration.NewWriterSaver(fh, logger, format)
	if err != nil {
		_ = fh.Close()
		return err
	}
	if err := saver.Save(config); err != nil {
		_ = fh.Close()
		return err
	}
	if err := fh.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, hostKeyFile); err != nil {
		return fmt.Errorf("failed to rename temporary file %s to %s (%w)", tmpFile, configFile, err)
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

func dumpConfig(writer io.Writer, loggerFactory log.LoggerFactory, config *configuration.AppConfig) error {
	configLogger, err := loggerFactory.Make(
		config.Log,
		"config",
	)
	if err != nil {
		return err
	}

	saver, err := configuration.NewWriterSaver(writer, configLogger, configuration.FormatYAML)
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
		"config",
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
