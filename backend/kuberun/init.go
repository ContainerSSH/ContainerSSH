package kuberun

import (
	"context"
	"fmt"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/config"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

func createSession(sessionId string, username string, appConfig *config.AppConfig) (backend.Session, error) {
	logrus.Trace(fmt.Sprintf("Initializing Kubernetes backend"))
	connectionConfig := restclient.Config{
		Host:    appConfig.KubeRun.Connection.Host,
		APIPath: appConfig.KubeRun.Connection.APIPath,
		ContentConfig: restclient.ContentConfig{
			GroupVersion:         &v1.SchemeGroupVersion,
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
		Username:        appConfig.KubeRun.Connection.Username,
		Password:        appConfig.KubeRun.Connection.Password,
		BearerToken:     appConfig.KubeRun.Connection.BearerToken,
		BearerTokenFile: appConfig.KubeRun.Connection.BearerTokenFile,
		Impersonate:     restclient.ImpersonationConfig{},
		TLSClientConfig: restclient.TLSClientConfig{
			Insecure:   appConfig.KubeRun.Connection.Insecure,
			ServerName: appConfig.KubeRun.Connection.ServerName,
			CertFile:   appConfig.KubeRun.Connection.CertFile,
			KeyFile:    appConfig.KubeRun.Connection.KeyFile,
			CAFile:     appConfig.KubeRun.Connection.CAFile,
			CertData:   []byte(appConfig.KubeRun.Connection.CertData),
			KeyData:    []byte(appConfig.KubeRun.Connection.KeyData),
			CAData:     []byte(appConfig.KubeRun.Connection.CAData),
		},
		UserAgent: "ContainerSSH",
		QPS:       appConfig.KubeRun.Connection.QPS,
		Burst:     appConfig.KubeRun.Connection.Burst,
		Timeout:   appConfig.KubeRun.Connection.Timeout,
	}

	cli, err := kubernetes.NewForConfig(&connectionConfig)
	if err != nil {
		return nil, err
	}

	restClient, err := restclient.RESTClientFor(&connectionConfig)
	if err != nil {
		return nil, err
	}

	session := &kubeRunSession{}
	session.sessionId = sessionId
	session.username = username
	session.env = map[string]string{}
	session.pty = false
	session.pod = nil
	session.client = cli
	session.ctx = context.Background()
	session.exitCode = -1
	session.config = appConfig.KubeRun
	session.restClient = restClient
	session.connectionConfig = connectionConfig

	return session, nil
}

type sizeQueue struct {
	resizeChan chan remotecommand.TerminalSize
}

func (s *sizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-s.resizeChan
	if !ok {
		return nil
	}
	return &size
}

type kubeRunSession struct {
	username          string
	sessionId         string
	env               map[string]string
	pty               bool
	pod               *v1.Pod
	exitCode          int32
	ctx               context.Context
	client            *kubernetes.Clientset
	config            config.KubeRunConfig
	terminalSizeQueue sizeQueue
	restClient        *restclient.RESTClient
	connectionConfig  restclient.Config
}

func Init(registry *backend.Registry) {
	kubeRunBackend := backend.Backend{}
	kubeRunBackend.Name = "kuberun"
	kubeRunBackend.CreateSession = createSession
	registry.Register(kubeRunBackend)
}
