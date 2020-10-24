package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	auditFactory "github.com/containerssh/containerssh/audit/factory"
	"github.com/containerssh/containerssh/auth"
	"github.com/containerssh/containerssh/backend"
	"github.com/containerssh/containerssh/backend/dockerrun"
	"github.com/containerssh/containerssh/backend/kuberun"
	backendMetrics "github.com/containerssh/containerssh/backend/metrics"
	configurationClient "github.com/containerssh/containerssh/config/client"
	"github.com/containerssh/containerssh/config/loader"
	"github.com/containerssh/containerssh/config/util"
	"github.com/containerssh/containerssh/geoip/dummy"
	"github.com/containerssh/containerssh/geoip/oschwald"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/log/writer"
	"github.com/containerssh/containerssh/metrics"
	metricsServer "github.com/containerssh/containerssh/metrics/server"
	"github.com/containerssh/containerssh/ssh"
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

	appConfig, err := util.GetDefaultConfig()
	if err != nil {
		logger.CriticalF("error getting default config (%v)", err)
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
			logger.EmergencyF("error loading config file (%v)", err)
			os.Exit(1)
		}
		appConfig, err = util.Merge(fileAppConfig, appConfig)
		if err != nil {
			logger.EmergencyF("error merging config (%v)", err)
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
			logger.EmergencyF("missing LICENSE.md, cannot print license information")
			os.Exit(1)
		}
		fmt.Println(string(data))
		fmt.Println("")
		data, err = ioutil.ReadFile("NOTICE.md")
		if err != nil {
			logger.EmergencyF("missing NOTICE.md, cannot print third party license information")
			os.Exit(1)
		}
		fmt.Println(string(data))
		fmt.Println("")
	}

	if dumpConfig || licenses {
		return
	}

	geoIpLookupProvider, err := oschwald.New(appConfig.GeoIP.GeoIP2File)
	if err != nil {
		logger.WarningF("failed to load GeoIP2 database, falling back to dummy provider (%v)", err)
		geoIpLookupProvider = dummy.New()
	}
	metricCollector := metrics.New(geoIpLookupProvider)

	backendRegistry := InitBackendRegistry(metricCollector)

	authClient, err := auth.NewHttpAuthClient(appConfig.Auth, logger, metricCollector)
	if err != nil {
		logger.CriticalF("error creating auth HTTP client (%v)", err)
		os.Exit(1)
	}

	configClient, err := configurationClient.NewHttpConfigClient(appConfig.ConfigServer, logger, metricCollector)
	if err != nil {
		logger.EmergencyF("error creating config HTTP client (%s)", err)
		os.Exit(1)
	}

	audit, err := auditFactory.Get(appConfig.Audit, logger)
	if err != nil {
		logger.EmergencyF("failed to create audit backend (%s)", err)
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
		audit,
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
