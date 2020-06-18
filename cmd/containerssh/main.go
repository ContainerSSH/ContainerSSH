package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/janoszen/containerssh/auth"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/backend/dockerrun"
	"github.com/janoszen/containerssh/backend/kuberun"
	configurationClient "github.com/janoszen/containerssh/config/client"
	"github.com/janoszen/containerssh/config/loader"
	"github.com/janoszen/containerssh/config/util"
	"github.com/janoszen/containerssh/ssh"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

func InitBackendRegistry() *backend.Registry {
	registry := backend.NewRegistry()
	dockerrun.Init(registry)
	kuberun.Init(registry)
	return registry
}

func main() {
	log.SetLevel(log.TraceLevel)

	backendRegistry := InitBackendRegistry()
	appConfig, err := util.GetDefaultConfig()
	if err != nil {
		log.Fatalf("Error getting default config (%s)", err)
		return
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
			log.Fatal(fmt.Sprintf("Error loading config file (%s)", err))
		}
		appConfig, err = util.Merge(fileAppConfig, appConfig)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error merging config (%s)", err))
		}
	}

	if dumpConfig {
		err := loader.Write(appConfig, os.Stdout)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error dumping config (%s)", err))
		}
	}

	if licenses {
		fmt.Println("# The ContainerSSH license")
		fmt.Println("")
		data, err := ioutil.ReadFile("LICENSE.md")
		if err != nil {
			log.Fatalf("Missing LICENSE.md, cannot print license information")
		}
		fmt.Println(string(data))
		fmt.Println("")
		data, err = ioutil.ReadFile("NOTICE.md")
		if err != nil {
			log.Fatalf("Missing NOTICE.md, cannot print third party license information")
		}
		fmt.Println(string(data))
		fmt.Println("")
	}

	if dumpConfig || licenses {
		return
	}

	authClient, err := auth.NewHttpAuthClient(appConfig.Auth)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating auth HTTP client (%s)", err))
		return
	}

	configClient, err := configurationClient.NewHttpConfigClient(appConfig.ConfigServer)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating config HTTP client (%s)", err))
		return
	}

	sshServer, err := ssh.NewServer(
		appConfig,
		authClient,
		backendRegistry,
		configClient,
	)
	if err != nil {
		log.Fatalf("failed to create SSH server (%v)", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	errChannel := make(chan error)
	go func() {
		err = sshServer.Run(ctx)
		if err != nil {
			errChannel <- err
		} else {
			errChannel <- nil
		}
	}()

	select {
	case <-sigs:
		log.Infof("received exit signal, stopping SSH server")
		cancel()
	case <-ctx.Done():
	case err = <-errChannel:
		cancel()
		log.Fatalf("failed to run SSH server (%v)", err)
	}
}
