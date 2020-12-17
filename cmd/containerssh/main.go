package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

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

type flags struct {
	configFile string
	dumpConfig bool
	licenses   bool
}

func parseFlags() (flags flags) {
	flag.StringVar(
		&flags.configFile,
		"config",
		"",
		"Configuration file to load (has to end in .yaml, .yml, or .json)",
	)
	flag.BoolVar(
		&flags.dumpConfig,
		"dump-config",
		false,
		"Dump configuration and exit",
	)
	flag.BoolVar(
		&flags.licenses,
		"licenses",
		false,
		"Print license information",
	)
	flag.Parse()

	return
}

func initBackendRegistry(metric *metrics.MetricCollector) *backend.Registry {
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
	logger := log.NewLoggerPipeline(logConfig, logWriter)

	flags := parseFlags()

	if flags.licenses {
		fmt.Println("# The ContainerSSH license")
		fmt.Println("")

		for _, f := range []string{"LICENSE.md", "NOTICE.md"} {
			data, err := ioutil.ReadFile(f)
			if err != nil {
				logger.EmergencyF("Missing %s, cannot print license information", f)
				os.Exit(1)
			}
			fmt.Println(string(data))
			fmt.Println("")
		}

		return
	}

	appConfig, err := util.GetDefaultConfig()
	if err != nil {
		logger.CriticalF("error getting default config (%v)", err)
		os.Exit(2)
	}

	if flags.configFile != "" {
		fileAppConfig, err := loader.LoadFile(flags.configFile)
		if err != nil {
			logger.EmergencyF("error loading config file (%v)", err)
			os.Exit(3)
		}
		appConfig, err = util.Merge(fileAppConfig, appConfig)
		if err != nil {
			logger.EmergencyF("error merging config (%v)", err)
			os.Exit(4)
		}
	}

	if flags.dumpConfig {
		err := loader.Write(appConfig, os.Stdout)
		if err != nil {
			logger.EmergencyF("error dumping config (%v)", err)
			os.Exit(5)
		}
		return
	}

	geoIPLookupProvider, err := oschwald.New(appConfig.GeoIP.GeoIP2File)
	if err != nil {
		logger.WarningF("failed to load GeoIP2 database, falling back to dummy provider (%v)", err)
		geoIPLookupProvider = dummy.New()
	}
	metricCollector := metrics.New(geoIPLookupProvider)

	backendRegistry := initBackendRegistry(metricCollector)

	authClient, err := auth.NewHttpAuthClient(appConfig.Auth, logger, metricCollector)
	if err != nil {
		logger.CriticalF("error creating auth HTTP client (%v)", err)
		os.Exit(6)
	}

	configClient, err := configurationClient.NewHttpConfigClient(appConfig.ConfigServer, logger, metricCollector)
	if err != nil {
		logger.EmergencyF(fmt.Sprintf("Error creating config HTTP client (%s)", err))
		os.Exit(7)
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
		os.Exit(8)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	sshErrChannel := make(chan error)
	go func() {
		sshErrChannel <- sshServer.Run(ctx)
	}()

	metricsErrChannel := make(chan error)
	if appConfig.Metrics.Enable {
		go func() {
			metricsSrv := metricsServer.New(
				appConfig.Metrics,
				metricCollector,
				logger,
			)
			metricsErrChannel <- metricsSrv.Run(ctx)
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
