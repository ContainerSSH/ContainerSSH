package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/janoszen/containerssh/auth"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/backend/dockerrun"
	"github.com/janoszen/containerssh/backend/kuberun"
	backendMetrics "github.com/janoszen/containerssh/backend/metrics"
	configurationClient "github.com/janoszen/containerssh/config/client"
	"github.com/janoszen/containerssh/config/loader"
	"github.com/janoszen/containerssh/config/util"
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/log/writer"
	"github.com/janoszen/containerssh/metrics"
	metricsServer "github.com/janoszen/containerssh/metrics/server"
	"github.com/janoszen/containerssh/ssh"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

func InitBackendRegistry(metric *metrics.MetricCollector) *backend.Registry {
	backendMetrics.Init(metric)
	registry := backend.NewRegistry()
	dockerrun.Init(registry, metric)
	kuberun.Init(registry, metric)
	return registry
}

func main() {
	logConfig, err := log.NewConfig(log.LevelInfoString)
	if err != nil {
		panic(err)
	}
	logWriter := writer.NewJsonLogWriter()
	var logger log.Logger
	logger = log.NewLoggerPipeline(logConfig, logWriter)
	metricCollector := metrics.New()

	backendRegistry := InitBackendRegistry(metricCollector)
	appConfig, err := util.GetDefaultConfig()
	if err != nil {
		logger.CriticalF("Error getting default config (%s)", err)
		os.Exit(1)
	}

	configFile := ""
	dumpConfig := false
	licenses := false
	generateHostKeys := false
	flag.StringVar(
		&configFile,
		"config",
		"",
		"Configuration file to load (has to end in .yaml, .yml, or .json)",
	)
	flag.BoolVar(
		&dumpConfig,
		"dump-config",
		false,
		"Dump configuration and exit",
	)
	flag.BoolVar(
		&licenses,
		"licenses",
		false,
		"Print license information",
	)
	flag.BoolVar(
		&generateHostKeys,
		"generate-host-keys",
		false,
		"Generate host keys if not present and exit",
	)
	flag.Parse()

	if configFile != "" {
		fileAppConfig, err := loader.LoadFile(configFile)
		if err != nil {
			logger.EmergencyF("Error loading config file (%v)", err)
			os.Exit(1)
		}
		appConfig, err = util.Merge(fileAppConfig, appConfig)
		if err != nil {
			logger.EmergencyF("Error merging config (%v)", err)
			os.Exit(1)
		}
	}

	if dumpConfig {
		err := loader.Write(appConfig, os.Stdout)
		if err != nil {
			logger.EmergencyF("error dumping config (%v)", err)
			os.Exit(1)
		}
	}

	if licenses {
		fmt.Println("# The ContainerSSH license")
		fmt.Println("")
		data, err := ioutil.ReadFile("LICENSE.md")
		if err != nil {
			logger.EmergencyF("Missing LICENSE.md, cannot print license information")
			os.Exit(1)
		}
		fmt.Println(string(data))
		fmt.Println("")
		data, err = ioutil.ReadFile("NOTICE.md")
		if err != nil {
			logger.EmergencyF("Missing NOTICE.md, cannot print third party license information")
			os.Exit(1)
		}
		fmt.Println(string(data))
		fmt.Println("")
	}

	if dumpConfig || licenses {
		return
	}

	authClient, err := auth.NewHttpAuthClient(appConfig.Auth, logger, metricCollector)
	if err != nil {
		logger.CriticalF("error creating auth HTTP client (%v)", err)
		os.Exit(1)
	}

	configClient, err := configurationClient.NewHttpConfigClient(appConfig.ConfigServer, logger, metricCollector)
	if err != nil {
		logger.EmergencyF(fmt.Sprintf("Error creating config HTTP client (%s)", err))
		os.Exit(1)
	}

	sshServer, err := ssh.NewServer(
		appConfig,
		authClient,
		backendRegistry,
		configClient,
		logger,
		logWriter,
		metricCollector,
	)
	if err != nil {
		logger.EmergencyF("failed to create SSH server (%v)", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sshErrChannel := make(chan error)
	go func() {
		err = sshServer.Run(ctx)
		if err != nil {
			sshErrChannel <- err
		} else {
			sshErrChannel <- nil
		}
	}()

	metricsErrChannel := make(chan error)
	if appConfig.Metrics.Enable {
		go func() {
			metricsSrv := metricsServer.New(
				appConfig.Metrics,
				metricCollector,
				logger,
			)
			err := metricsSrv.Run(ctx)
			if err != nil {
				metricsErrChannel <- err
			} else {
				metricsErrChannel <- nil
			}
		}()
	} else {
		metricsErrChannel <- nil
	}

	select {
	case <-sigs:
		logger.InfoF("received exit signal, stopping SSH server")
		cancel()
	case <-ctx.Done():
	case err = <-metricsErrChannel:
		cancel()
		if err != nil {
			logger.EmergencyF("failed to run HTTP server (%v)", err)
		}
	case err = <-sshErrChannel:
		cancel()
		if err != nil {
			logger.EmergencyF("failed to run SSH server (%v)", err)
		}
	}
}
