package test

import (
    "archive/tar"
    "compress/gzip"
    "context"
    "embed"
    "fmt"
    "io"
    "strconv"
    "strings"
    "testing"
    "time"

    "github.com/docker/docker/api/types"
    containerType "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
    "github.com/docker/go-connections/nat"
)

func dockerClient(t *testing.T) *client.Client {
	t.Helper()
	cli, err := client.NewClientWithOpts()
	if err != nil {
		t.Fatalf("failed to obtain Docker client (%v)", err)
	}
	cli.NegotiateAPIVersion(context.Background())
	return cli
}

// containerFromBuild starts a container from a local image build.
//
// - imageTag specifies the name the built image should be tagged as.
// - files specifies the list of local files to send as part of the build context.
// - cmd specifies the command line in execv format if any.
// - env specifies the environment variables in the VAR=VALUE format.
// - ports specifies the port mappings. The key is the container port, the value is the host port.
//   If the host port is left empty it is automatically mapped and can be retrieved later.
func containerFromBuild(
	t *testing.T,
	imageTag string,
	files map[string][]byte,
	cmd []string,
	env []string,
	ports map[string]string,
) container {
	t.Helper()
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
	cnt.inspect()

	return cnt
}

// containerFromBuild starts a container from an image pulled from a registry
//
// - image specifies the image tag to pull.
// - cmd specifies the command line in execv format if any.
// - env specifies the environment variables in the VAR=VALUE format.
// - ports specifies the port mappings. The key is the container port, the value is the host port.
//   If the host port is left empty it is automatically mapped and can be retrieved later.
func containerFromPull(
	t *testing.T,
	image string,
	cmd []string,
	env []string,
	ports map[string]string,
) container {
	t.Helper()
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
	cnt.inspect()

	return cnt
}

type container interface {
	// id returns the container ID of the container.
	id() string
	// port returns the host port the containerPort has been mapped to.
	port(containerPort string) int
	// extractFile extracts a single file from the container and returns the contents.
	extractFile(fileName string) []byte
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

func (d *dockerContainer) extractFile(fileName string) []byte {
	d.t.Helper()
	keyReader, _, err := d.client.CopyFromContainer(context.Background(), d.containerID, fileName)
	if err != nil {
		d.t.Fatalf("failed to retrieve file %s key from the container (%v)", fileName, err)
	}
	defer func() {
		_ = keyReader.Close()
	}()

	tarReader := tar.NewReader(keyReader)
	if _, err := tarReader.Next(); err != nil {
		d.t.Fatalf("failed to get next file in TAR archive from container (%v)", err)
	}

	data, err := io.ReadAll(tarReader)
	if err != nil {
		d.t.Fatalf("failed to parse TAR archive from container (%v)", err)
	}
	return data
}

func (d *dockerContainer) port(containerPort string) int {
	hostPortString := d.ports[containerPort]
	hostPort, err := strconv.ParseInt(hostPortString, 10, 64)
	if err != nil {
		panic(fmt.Errorf("BUG: failed to parse container %s host port %s (%w)", d.image, hostPortString, err))
	}
	if hostPort < 1 || hostPort > 65536 {
		panic(fmt.Errorf(
			"BUG: Docker daemon returned invalid host port number for %s: %d",
			containerPort,
			hostPort,
		))
	}
	return int(hostPort)
}

type testLogWriter struct {
	t *testing.T
}

func (n2 testLogWriter) Write(p []byte) (n int, err error) {
	n2.t.Helper()
	n2.t.Logf("%s", p)

	return len(p), nil
}

func (d *dockerContainer) id() string {
	return d.containerID
}

func (d *dockerContainer) pull() {
	d.t.Helper()
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
	d.t.Helper()
	d.t.Logf("Creating container from %s...", d.image)
	hostConfig := &containerType.HostConfig{
		AutoRemove:   true,
		PortBindings: map[nat.Port][]nat.PortBinding{},
	}
	for containerPort, hostPort := range d.ports {
		portString := nat.Port(containerPort)
		if hostPort == "" {
			hostConfig.PortBindings[portString] = []nat.PortBinding{
				{
					HostIP: "127.0.0.1",
				},
			}
		} else {
			hostConfig.PortBindings[portString] = []nat.PortBinding{
				{
					HostIP:   "127.0.0.1",
					HostPort: hostPort,
				},
			}
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
	d.t.Helper()
	d.t.Logf("Starting container %s...", d.containerID)
	if err := d.client.ContainerStart(context.Background(), d.containerID, types.ContainerStartOptions{}); err != nil {
		d.t.Fatalf("failed to start container %s (%v)", d.containerID, err)
	}
}

func (d *dockerContainer) remove() {
	d.t.Helper()
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
	d.t.Helper()
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
				Size:    int64(len(fileContent)),
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
			Tags:       []string{d.image},
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

func (d *dockerContainer) inspect() {
	d.t.Helper()
	d.t.Logf("Waiting for the port mappings to come up...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for {
		d.t.Logf("Inspecting container %s...", d.containerID)
		inspectResult, err := d.client.ContainerInspect(ctx, d.containerID)
		if err != nil {
			d.t.Fatalf("Failed to inspect container (%v)", err)
		}
		if inspectResult.State.Health == nil || inspectResult.State.Health.Status == types.Healthy ||
			inspectResult.State.Health.Status == types.NoHealthcheck {
			if inspectResult.State.Running {
				if d.updatePortMappings(inspectResult) {
					return
				} else {
					d.t.Logf("Ports are not mapped yet...")
				}
			} else {
				d.t.Logf("Container is not running yet, it is %s...", inspectResult.State.Status)
			}
		} else {
			d.t.Logf("Container is not healthy yet, health is %s...", inspectResult.State.Health.Status)
		}
		select {
		case <-ctx.Done():
			d.t.Fatalf("Failed to inspect Minio container (timeout)")
		case <-time.After(time.Second):
		}
	}
}

func (d *dockerContainer) updatePortMappings(inspectResult types.ContainerJSON) bool {
	for port := range d.ports {
		if hostPorts, ok := inspectResult.NetworkSettings.Ports[nat.Port(port)]; ok {
			if len(hostPorts) > 0 {
				d.ports[port] = hostPorts[0].HostPort
			} else {
				return false
			}
		}
	}
	return true
}

func dockerBuildRootFiles(buildRoot embed.FS, startingDir string) map[string][]byte {
	files := dockerBuildRootDirFiles(buildRoot, startingDir)
	result := map[string][]byte{}
	for file, data := range files {
		result[strings.TrimPrefix(file, fmt.Sprintf("%s/", startingDir))] = data
	}
	return result
}

func dockerBuildRootDirFiles(buildRoot embed.FS, dir string) map[string][]byte {
	result := map[string][]byte{}
	fsEntries, err := buildRoot.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, fsEntry := range fsEntries {
		fullPath := dir + "/" + fsEntry.Name()
		if fsEntry.IsDir() {
			subDirFiles := dockerBuildRootDirFiles(buildRoot, fullPath)
			for fileName, fileContent := range subDirFiles {
				result[fileName] = fileContent
			}
		} else {
			data, err := buildRoot.ReadFile(fullPath)
			if err != nil {
				panic(err)
			}
			result[fullPath] = data
		}
	}
	return result
}
