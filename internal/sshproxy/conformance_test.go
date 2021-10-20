package sshproxy_test

import (
	"archive/tar"
	"context"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/geoip/dummy"
	"github.com/containerssh/containerssh/internal/metrics"
	"github.com/containerssh/containerssh/internal/sshserver"
	"github.com/containerssh/containerssh/internal/sshproxy"
	"github.com/containerssh/containerssh/internal/structutils"
	"github.com/containerssh/containerssh/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"golang.org/x/crypto/ssh"
)

func TestConformance(t *testing.T) {
	fingerprint := setUpBackendContainer(t)

	var factories = map[string]func(logger log.Logger) (sshserver.NetworkConnectionHandler, error){
		"sshproxy": func(logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			connectionID := sshserver.GenerateConnectionID()
			geoipProvider := dummy.New()
			cfg := config.SSHProxyConfig{}
			structutils.Defaults(&cfg)
			cfg.Server = "127.0.0.1"
			cfg.Port = 22222
			cfg.Username = "test"
			cfg.Password = "test"
			cfg.AllowedHostKeyFingerprints = []string{
				fingerprint,
			}
			cfg.HostKeyAlgorithms = []config.SSHKeyAlgo{
				config.SSHKeyAlgoSSHRSA,
			}
			collector := metrics.New(geoipProvider)
			return sshproxy.New(
				net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 2222,
					Zone: "",
				},
				connectionID,
				cfg,
				logger,
				collector.MustCreateCounter("backend_requests", "", ""),
				collector.MustCreateCounter("backend_failures", "", ""),
			)
		},
	}

	sshserver.RunConformanceTests(t, factories)
}

func setUpBackendContainer(t *testing.T) string {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancelFunc()
	cli, err := client.NewClientWithOpts()
	if err != nil {
		t.Fatalf("failed to create Docker client (%v)", err)
	}
	cli.NegotiateAPIVersion(ctx)
	pullContainerImage(ctx, t, cli)
	cnt := createContainer(ctx, t, cli)
	if err := cli.ContainerStart(ctx, cnt.ID, types.ContainerStartOptions{}); err != nil {
		t.Fatalf("failed to start container (%v)", err)
	}

	private := getHostKeyFromContainer(ctx, t, cli, cnt)
	fingerprint := ssh.FingerprintSHA256(private.PublicKey())
	return fingerprint
}

func createContainer(
	ctx context.Context,
	t *testing.T,
	cli *client.Client,
) container.ContainerCreateCreatedBody {
	cnt, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Cmd: []string{
				"/bin/bash",
				"-c",
				"mkdir /run/sshd && useradd -s /bin/bash -m test && echo 'test:test' | chpasswd && /usr/sbin/sshd -D -o AcceptEnv=*",
			},
			Image: "containerssh/containerssh-guest-image",
			ExposedPorts: map[nat.Port]struct{}{
				"22": {},
			},
		},
		&container.HostConfig{
			AutoRemove: false,
			Tmpfs: map[string]string{
				"/tmp": "",
				"/run": "",
			},
			PortBindings: map[nat.Port][]nat.PortBinding{
				"22": {
					{
						HostIP:   "127.0.0.1",
						HostPort: "22222",
					},
				},
			},
		},
		nil,
		nil,
		"",
	)
	if err != nil {
		t.Fatalf("failed to create backing SSH server container (%v)", err)
	}
	t.Cleanup(
		func() {
			t.Log("Removing backing container...")
			_ = cli.ContainerRemove(
				context.Background(), cnt.ID, types.ContainerRemoveOptions{
					Force: true,
				},
			)
		},
	)
	return cnt
}

func pullContainerImage(ctx context.Context, t *testing.T, cli *client.Client) {
	reader, err := cli.ImagePull(ctx, "docker.io/containerssh/containerssh-guest-image", types.ImagePullOptions{})
	if err != nil {
		t.Fatalf("failed to pull containerssh/containerssh-guest-image (%v)", err)
	}
	if _, err := ioutil.ReadAll(reader); err != nil {
		t.Fatalf("failed to pull containerssh/containerssh-guest-image (%v)", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("failed to pull containerssh/containerssh-guest-image (%v)", err)
	}
}

func getHostKeyFromContainer(
	ctx context.Context,
	t *testing.T,
	cli *client.Client,
	cnt container.ContainerCreateCreatedBody,
) ssh.Signer {
	keyReader, _, err := cli.CopyFromContainer(ctx, cnt.ID, "/etc/ssh/ssh_host_rsa_key")
	if err != nil {
		t.Fatalf("failed to retrieve the SSH host key from the container (%v)", err)
	}
	defer func() {
		_ = keyReader.Close()
	}()

	tarReader := tar.NewReader(keyReader)
	if _, err := tarReader.Next(); err != nil {
		t.Fatalf("failed to get next file in TAR archive from container (%v)", err)
	}

	data, err := ioutil.ReadAll(tarReader)
	if err != nil {
		t.Fatalf("failed to parse TAR archive from container (%v)", err)
	}

	private, err := ssh.ParsePrivateKey(data)
	if err != nil {
		t.Fatalf("failed to parse SSH host key (%v)", err)
	}
	return private
}
