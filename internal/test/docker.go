package test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	containerType "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func dockerClient(t *testing.T) *client.Client {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		t.Fatalf("failed to obtain Docker client (%v)", err)
	}
	cli.NegotiateAPIVersion(context.Background())
	return cli
}

func containerFromBuild(
	t *testing.T,
	imageTag string,
	files map[string][]byte,
	cmd []string,
	env []string,
	ports map[string]string,
) container {
	t.Logf("Creating and starting a container from local build...")
	cnt := &dockerContainer{
		client: dockerClient(t),
		files:  files,
		image:  imageTag,
		t:      t,
		cmd:    cmd,
		env:    env,
		ports:  ports,
	}

	cnt.build()
	cnt.create()
	t.Cleanup(cnt.remove)
	cnt.start()

	return cnt
}

func containerFromPull(
	t *testing.T,
	image string,
	cmd []string,
	env []string,
	ports map[string]string,
) container {
	t.Logf("Creating and starting a container from %s...", image)
	cnt := &dockerContainer{
		client: dockerClient(t),
		image:  image,
		t:      t,
		cmd:    cmd,
		env:    env,
		ports:  ports,
	}

	cnt.pull()
	cnt.create()
	t.Cleanup(cnt.remove)
	cnt.start()

	return cnt
}

type container interface {
	id() string
}

type dockerContainer struct {
	t      *testing.T
	client *client.Client

	containerID string
	image       string
	cmd         []string
	env         []string
	ports       map[string]string
	files       map[string][]byte
}

type testLogWriter struct {
	t *testing.T
}

func (n2 testLogWriter) Write(p []byte) (n int, err error) {
	n2.t.Logf("%s", p)

	return len(p), nil
}

func (d *dockerContainer) id() string {
	return d.containerID
}

func (d *dockerContainer) pull() {
	d.t.Logf("Pulling image %s...", d.image)
	reader, err := d.client.ImagePull(context.Background(), d.image, types.ImagePullOptions{})
	if err != nil {
		d.t.Fatalf("failed to pull container image %s (%v)", d.image, err)
	}
	if _, err := io.Copy(&testLogWriter{t: d.t}, reader); err != nil {
		d.t.Fatalf("failed to stream logs from image pull (%v)", err)
	}
}

func (d *dockerContainer) create() {
	d.t.Logf("Creating container from %s...", d.image)
	hostConfig := &containerType.HostConfig{
		AutoRemove:   true,
		PortBindings: map[nat.Port][]nat.PortBinding{},
	}
	for containerPort, hostPort := range d.ports {
		portString := nat.Port(containerPort)
		hostConfig.PortBindings[portString] = []nat.PortBinding{
			{
				HostIP:   "127.0.0.1",
				HostPort: hostPort,
			},
		}
	}
	resp, err := d.client.ContainerCreate(
		context.Background(),
		&containerType.Config{
			Image: d.image,
			Cmd:   d.cmd,
			Env:   d.env,
		},
		hostConfig,
		nil,
		nil,
		"",
	)
	if err != nil {
		d.t.Fatalf("failed to create %s container (%v)", d.image, err)
	}
	d.containerID = resp.ID
	d.t.Logf("Container has ID %s...", d.containerID)
}

func (d *dockerContainer) start() {
	d.t.Logf("Starting container %s...", d.containerID)
	if err := d.client.ContainerStart(context.Background(), d.containerID, types.ContainerStartOptions{}); err != nil {
		d.t.Fatalf("failed to start container %s (%v)", d.containerID, err)
	}
}

func (d *dockerContainer) remove() {
	d.t.Logf("Removing container ID %s...", d.containerID)
	if err := d.client.ContainerRemove(context.Background(), d.containerID, types.ContainerRemoveOptions{
		RemoveVolumes: false,
		RemoveLinks:   false,
		Force:         true,
	}); err != nil {
		d.t.Fatalf("failed to remove container %s (%v)", d.containerID, err)
	}
}

func (d *dockerContainer) build() {
	d.t.Logf("Building local image...")

	buildContext, buildContextWriter := io.Pipe()

	go func() {
		d.t.Logf("Compiling build context...")
		defer func() {
			d.t.Logf("Build context compiled.")
			_ = buildContextWriter.Close()
		}()

		gzipWriter := gzip.NewWriter(buildContextWriter)
		defer func() {
			_ = gzipWriter.Close()
		}()

		tarWriter := tar.NewWriter(gzipWriter)
		defer func() {
			_ = tarWriter.Close()
		}()

		for filePath, fileContent := range d.files {
			header := &tar.Header{
				Name:    filePath,
				Size: int64(len(fileContent)),
				Mode:    0755,
				ModTime: time.Now(),
			}

			if err := tarWriter.WriteHeader(header); err != nil {
				d.t.Logf("Failed to write build context file %s header (%v)", filePath, err)
				return
			}

			if _, err := tarWriter.Write(fileContent); err != nil {
				d.t.Logf("Failed to write build context file %s (%v)", filePath, err)
			}
		}
	}()

	imageBuildResponse, err := d.client.ImageBuild(
		context.Background(),
		buildContext,
		types.ImageBuildOptions{
			Tags:           []string{d.image},
			Dockerfile: "Dockerfile",
		},
	)
	if err != nil {
		d.t.Fatalf("Failed to build local image (%v)", err)
	}

	d.t.Logf("Reading build log...")
	if _, err := io.Copy(&testLogWriter{t: d.t}, imageBuildResponse.Body); err != nil {
		d.t.Fatalf("Failed to read image build log (%v)", err)
	}
	if err := imageBuildResponse.Body.Close(); err != nil {
		d.t.Fatalf("Failed to close image build log (%v)", err)
	}
}
