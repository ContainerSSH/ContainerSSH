# ContainerSSH: An SSH server that launches containers

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/janoszen/containerssh/goreleaser)](https://github.com/janoszen/containerssh/actions)
[![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/janoszen/containerssh)](http://hub.docker.com/r/janoszen/containerssh)
[![Go Report Card](https://goreportcard.com/badge/github.com/janoszen/containerssh)](https://goreportcard.com/report/github.com/janoszen/containerssh)
[![LGTM Alerts](https://img.shields.io/lgtm/alerts/github/janoszen/containerssh)](https://lgtm.com/projects/g/janoszen/containerssh/)
[![GitHub](https://img.shields.io/github/license/janoszen/containerssh)](LICENSE.md)

This is a Proof of Concept SSH server written in Go that sends any shell directly into a Docker container or Kubernetes
pod instead of launching it on a local machine. It uses an HTTP microservice as an authentication endpoint for SSH
connections.

## What is this?

This is an **SSH server that launches containers for every incoming connection**. You can run it on the host or in a
container. It needs two things: an authentication server and access to your container environment.

![Animation: SSH-ing into this SSH server lands you in a container where you can't access the network and you can't see any processes.](https://pasztor.at/assets/img/ssh-in-action.gif)

## Quick start

This is a quick start guide to get a test server up and running in less than 5 minutes with
[docker-compose](https://docs.docker.com/compose/).

To run it grab all files from the [example](example/) directory and run `docker-compose build` followed by 
`docker-compose up` in that directory. This will run the SSH server on your local machine on port 2222. You can log in
with any password using the user "foo" to get an Ubuntu image and "busybox" to get a Busybox image. 

## Use cases

- **Web hosting:** Imagine user A has access to site X and Y, user B has access to site Y and Z. You can use
  ContainerSSH to mount the appropriate sites for the SSH session.
- **Practicing environments:** Launch dummy containers for practice environment.
- **Honeypot:** Let attackers into an enclosed environment and observe them.

## How does it work?

```
+------+        +--------------+   2.   +-------------------+
|      |        |              | -----> |    Auth server    |
|      |        |              |        +-------------------+
|      |        |              |   
|      |   1.   |              |   3.   +-------------------+
| User | -----> | ContainerSSH | -----> |   Config server   |
|      |        |              |        +-------------------+
|      |        |              |   
|      |        |              |   4.   +-------------------+
|      |        |              | -----> | Container Backend |
+------+        +--------------+        +-------------------+
```

1. The user opens an SSH connection to ContainerSSH.
2. ContainerSSH calls the authentication server with the users username and password/pubkey to check if its valid.
3. ContainerSSH calls the config server to obtain backend location and configuration (if configured)
4. ContainerSSH calls the container backend to launch the container with the
   specified configuration. All input from the user is sent directly to the backend, output from the container is sent
   to the user.
   
> **Curious?** [Learn more about how this works in my blog post.](https://pasztor.at/blog/ssh-direct-to-docker)

## Installing

You can run ContainerSSH directly in containers by using the
[janoszen/containerssh](https://hub.docker.com/repository/docker/janoszen/containerssh) image name. Check the
[docker-compose.yaml example](example/docker-compose.yaml) for details how to set it up.

### Can I install it without containers?

Yes, but for now you will have to build it yourself. Once it hits the first stable release binary releases will be
provided.

## Building

The project can be built by running `go build cmd/containerssh/main.go` or using [goreleaser](https://goreleaser.com/)
by running `goreleaser build --snapshot --rm-dist`.

## Configuring

Before you can run ContainerSSH you will need to create a configuration file. The minimal configuration file looks like
this:

```yaml
ssh:
  hostkeys:
    # Generate a host key with ssh-keygen -t rsa
    - /path/to/your/host/key
auth:
  # See auth server below
  url: http://your-auth-server/
  password: true # Perform password authentication
  pubkey: false # Perform public key authentication
```

Note that the config file must end in `.yml`, `.yaml`, or `.json`. You can dump the entire configuration file using
`./containerssh --dump-config`

## Running

You can run the ContainerSSH server using the following command line:

```
./containerssh --config your/config-file.yaml
```

## Implementing an authentication server

ContainerSSH does not know your users and their passwords. Therefore, it calls out to a microservice that you have to
provide so it can verify the users, passwords and SSH keys. You will have to provide the microservice URL in the
configuration.

For password authentication ContainerSSH will call out to `http://your-auth-server/password` with the following request
body. The password is base64 encoded to transfer special characters properly.

```json
{
    "user": "username",
    "remoteAddress": "127.0.0.1:1234",
    "sessionId": "A base64 SSH session ID",
    "passwordBase64": "Base 64 password"
}
```

The public key auth ContainerSSH will call out to `http://your-auth-server/pubkey` in the following format.

```json
{
    "user": "username",
    "remoteAddress": "127.0.0.1:1234",
    "sessionId": "A base64 SSH session ID",
    "publicKeyBase64": "Base 64 public key in SSH wire format"
}
```

The public key is provided in the SSH wire format in base64 encoding.

Both endpoints need to respond with the following JSON:

```json
{
  "success": true
}
```

> **Tip** You can find the source code for a test authentication and configuration server written in Go
> [in this repository](cmd/testAuthConfigServer/main.go)

## Backend selection

ContainerSSH is built to support multiple backends. The backend can be changed in the configuration file:

```yaml
# change to `kuberun` to talk to Kubernetes
backend: dockerrun
```

## Configuration server

You have the option to dynamically change the configuration based on the username by providing a config server URL:

```yaml
configserver:
  url: http://your-config-server-url/
```

The config server will receive a request in following format:

```json
{
  "username":"ssh username",
  "sessionId": "ssh session ID"
}
```

Your application will have to respond in the following format:

```json
{
  "config": {
    // Provide a partial configuration here 
  }
}
```

You can view the full configuration structure in YAML format by running `./containerssh --dump-config`. Note that your
config server must respond in JSON format.

Some configuration values cannot be overridden from the config server. These are the ones that get used before the
connection is established, but the backends also override a few, such as `attachstdio`.

## The `dockerrun` backend

The `dockerrun` backend launches a container using the Docker API

### Providing certificates

Docker sockets allow connections over TCP with TLS encryption. You can provide these TLS certificates embedded in the
YAML file in PEM format:

```yaml
dockerrun:
  host: tcp://127.0.0.1:2376
  cacert: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  cert: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  key: |
    -----BEGIN RSA PRIVATE KEY-----
    ...
    -----END RSA PRIVATE KEY-----
```

### Changing the container image

The container image depends on the backend you are using. For `dockerrun` you can change the image in the config
file:

```yaml
dockerrun:
  config:
    container:
      image: your/image
``` 

### Detailed configuration

The full configuration at the time of writing are as described below. Keep in mind that the configuration structure
may change over time as they follow the
[Docker API](https://docs.docker.com/engine/api/v1.40/#operation/ContainerCreate).

```yaml
dockerrun:
    host: tcp://127.0.0.1:2375
    cacert: ""
    cert: ""
    key: ""
    config:
        container:
            hostname: ""
            domainname: ""
            user: ""
            # The "attach" options are overridden and cannot be configured
            #attachstdin: false
            #attachstdout: false
            #attachstderr: false
            exposedports: {}
            # The "tty" option depends on the requested SSH mode and cannot be configured
            #tty: false
            # The "stdin" options are also configured by the backend
            #openstdin: false
            #stdinonce: false
            # Env can be configured but will be overridden by the values provided via SSH
            env: []
            # CMD can be provided but will be overridden by the command sent via SSH
            cmd: []
            healthcheck: null
            argsescaped: false
            image: janoszen/containerssh-image
            volumes: {}
            workingdir: ""
            entrypoint: []
            networkdisabled: false
            macaddress: ""
            onbuild: []
            labels: {}
            stopsignal: ""
            stoptimeout: null
            shell: []
        host:
            binds: []
            containeridfile: ""
            logconfig:
                type: ""
                config: {}
            networkmode: ""
            portbindings: {}
            restartpolicy:
                name: ""
                maximumretrycount: 0
            autoremove: false
            volumedriver: ""
            volumesfrom: []
            capadd: []
            capdrop: []
            dns: []
            dnsoptions: []
            dnssearch: []
            extrahosts: []
            groupadd: []
            ipcmode: ""
            cgroup: ""
            links: []
            oomscoreadj: 0
            pidmode: ""
            privileged: false
            publishallports: false
            readonlyrootfs: false
            securityopt: []
            storageopt: {}
            tmpfs: {}
            utsmode: ""
            usernsmode: ""
            shmsize: 0
            sysctls: {}
            runtime: ""
            consolesize:
              - 0
              - 0
            isolation: ""
            resources:
                cpushares: 0
                memory: 0
                nanocpus: 0
                cgroupparent: ""
                blkioweight: 0
                blkioweightdevice: []
                blkiodevicereadbps: []
                blkiodevicewritebps: []
                blkiodevicereadiops: []
                blkiodevicewriteiops: []
                cpuperiod: 0
                cpuquota: 0
                cpurealtimeperiod: 0
                cpurealtimeruntime: 0
                cpusetcpus: ""
                cpusetmems: ""
                devices: []
                diskquota: 0
                kernelmemory: 0
                memoryreservation: 0
                memoryswap: 0
                memoryswappiness: null
                oomkilldisable: null
                pidslimit: 0
                ulimits: []
                cpucount: 0
                cpupercent: 0
                iomaximumiops: 0
                iomaximumbandwidth: 0
            mounts: []
            init: null
            initpath: ""
        network:
            endpointsconfig: {}
        containername: ""
```

## The `kuberun` backend

The kuberun backend runs a pod in a Kubernetes cluster and attaches to a container there.

### Running outside of Kubernetes

If you are running ContainerSSH outside of Kubernetes you will need the following configuration:

```yaml
kuberun:
    connection:
        host: your-kubernetes-api-server:6443
        cert: |
            -----BEGIN CERTIFICATE-----
            ...
            -----END CERTIFICATE-----
        key: |
            -----BEGIN RSA PRIVATE KEY-----
            ...
            -----END RSA PRIVATE KEY-----
        cacert: |
            -----BEGIN CERTIFICATE-----
            ...
            -----END CERTIFICATE-----
```

Alternatively you can use `cacertFile`, `keyFile` and `certFile` to point to files on the filesystem.

### Running inside a Kubernetes cluster

When you run inside of a Kubernetes cluster you can use the service account token:

```yaml
kuberun:
    connection:
        certFile: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
```

### Changing the container image

For the `kuberun` backend the container image can be changed by modifying the pod spec:

```yaml
kuberun:
    pod:
        namespace: default
        consoleContainerNumber: 0
        podSpec:
            volumes: []
            initcontainers: []
            containers:
              - name: shell
                image: janoszen/containerssh-image
```

Note: if you are running multiple containers you should specify the `consoleContainerNumber` parameter to indicate
which container you wish to attach to when an SSH session is opened.

### Detailed configuration

The full configuration looks as follows:

```yaml
kuberun:
    connection:
        host: kubernetes.docker.internal:6443
        path: /api
        username: docker-desktop
        password: ""
        insecure: false
        serverName: ""
        certFile: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        keyFile: ""
        cacertFile: ""
        cert: |
            -----BEGIN CERTIFICATE-----
            ...
            -----END CERTIFICATE-----
        key: |
            -----BEGIN RSA PRIVATE KEY-----
            ...
            -----END RSA PRIVATE KEY-----
        cacert: |
            -----BEGIN CERTIFICATE-----
            ...
            -----END CERTIFICATE-----
        bearerToken: ""
        bearerTokenFile: ""
        qps: 5
        burst: 10
        timeout: 0s
    pod:
        namespace: default
        # If you have multiple containers which container should the SSH session attach to?
        consoleContainerNumber: 0
        podSpec:
            volumes: []
            initcontainers: []
            containers:
              # The name doesn't matter
              - name: shell
                image: janoszen/containerssh-image
                # This may be overridden by the SSH client
                command: []
                args: []
                workingdir: ""
                ports: []
                envfrom: []
                # These will be populated from the SSH session but you can provide additional ones.
                env: []
                resources:
                    limits: {}
                    requests: {}
                volumemounts: []
                volumedevices: []
                livenessprobe: null
                readinessprobe: null
                startupprobe: null
                lifecycle: null
                terminationmessagepath: ""
                terminationmessagepolicy: ""
                imagepullpolicy: ""
                securitycontext: null
                # These 3 will be overridden based on the SSH session
                #stdin: false
                #stdinonce: false
                #tty: false
            ephemeralcontainers: []
            restartpolicy: ""
            terminationgraceperiodseconds: null
            activedeadlineseconds: null
            dnspolicy: ""
            nodeselector: {}
            serviceaccountname: ""
            deprecatedserviceaccount: ""
            automountserviceaccounttoken: null
            nodename: ""
            hostnetwork: false
            hostpid: false
            hostipc: false
            shareprocessnamespace: null
            securitycontext: null
            imagepullsecrets: []
            hostname: ""
            subdomain: ""
            affinity: null
            schedulername: ""
            tolerations: []
            hostaliases: []
            priorityclassname: ""
            priority: null
            dnsconfig: null
            readinessgates: []
            runtimeclassname: null
            enableservicelinks: null
            preemptionpolicy: null
            overhead: {}
            topologyspreadconstraints: []
        subsystems:
            # This will be used as `command` when the client asks for the SFTP subsystem.
            sftp: /usr/lib/openssh/sftp-server
    timeout: 1m0s
```

## Building a container image for ContainerSSH

ContainerSSH has no requirements as to the container image you are running apart from the fact that they need to be 
Linux containers.

If you wish to use SFTP you have to add an SFTP server (`apt install openssh-sftp-server` on Ubuntu) to the container
image and configure the path of the SFTP server correctly in your config.yaml. The sample image
`janoszen/containerssh-image` contains an SFTP server.

## FAQ

### Is ContainerSSH secure?

ContainerSSH depends on a number of libraries to achieve what it does. A security hole in any of the critical ones
could mean a compromise of your container environment, especially if you are using the `dockerrun` backend. (Docker
has no access control so a compromise means your whole host is compromised.)

### Is ContainerSSH production-ready?

No. ContainerSSH is very early in its development and has not undergone extensive testing yet. You should be careful
before deploying it into production.

### Does ContainerSSH delete containers after it is done?

ContainerSSH does its best to delete containers it creates. However, at this time there is no cleanup mechanism in case
it crashes.

### Do I need to run ContainerSSH as root?

No! In fact, you shouldn't! ContainerSSH is perfectly fine running as non-root as long as it has access to Kubernetes
or Docker. (Granted, access to the Docker socket means it could easily launch a root process on the host.)

### Does ContainerSSH support SFTP?

Yes, but your container image must contain an SFTP server binary and your config.yaml or config server must contain the
correct path for the server.

### Does ContainerSSH support SCP?

Not at this time.

### Does ContainerSSH support TCP port forwarding?

Not at this time. TCP port forwarding is done outside of a channel. At this time ContainerSSH launches one container
per SSH channel, so forwarding TCP ports would mean a complete overhaul of the entire architecture. In essence the
architecture would be changed to launch one container per session, not per channel, and use `exec` to launch a shell or
SFTP server for the channel.

However, as you might imagine that's a bit of a larger change and will need quite a bit of work.

### Does ContainerSSH support SSH agent forwarding?

Not at this time. SSH agent forwarding would require a separate binary agent within the container to proxy data.
This is similar to how TCP port forwarding works, except that the authentication agent requests are sent on a
per-channel basis. Additionally SSH agent forwarding is not documented well, it is proprietary to OpenSSH.
(The request type is `auth-agent-req@openssh.com`.)

### Does ContainerSSH support X11 forwarding?

Not at this time. X11 is sent over separate channels and would most probably need the overhaul that TCP port forwarding
requires. As X11 forwarding isn't use much any more it is unlikely that ContainerSSH will ever support it.

### Does ContainerSSH support forwarding signals?

Partially. The `dockerrun` backend supports it, the `kuberun` backend doesn't because Kubernetes itself doesn't.

### Does ContainerSSH support window resizing?

Yes.

### Does ContainerSSH support environment variable passing?

Yes.

### Does ContainerSSH support returning the exit status?

Partially. The `dockerrun` backend supports it, the `kuberun` backend &ldquo;does its best&rdquo; but has some
edge cases when the connection closes before the exit status can be obtained.

### Can ContainerSSH run exec into existing containers?

Not at this time. The architecture needs to solidify before such a feature is implemented.

### Can ContainerSSH deploy additional services, such as sidecar containers, etc?

ContainerSSH supports the entire Kubernetes pod specification so you can launch as many containers as you want in a
single pod. The Docker backend, however, does not support sidecar containers.

### Can I add metadata to my pods with the `kuberun` backend?

Not at this time. You may want to open up a feature request and detail your use case.

### Why is the `kuberun` backend so slow?

Kubernetes is built for scale. That means there are some tradeoffs in terms of responsiveness. This is not something
ContainerSSH can do anything about, it just takes a bit to launch a pod. You may want to fine-tune your Kubernetes 
cluster for responsiveness.

### Why is there no initial prompt with the `kuberun` backend?

This is a [known bug](https://github.com/janoszen/containerssh/issues/12). Unfortunately the `kuberun` backend was 
built by reverse engineering kubectl as there is no documentation whatsoever on how the attach functionality works on
pods. If you are good with Go you might want to help out here.

### Can I use my normal kubeconfig files?

Unfortunately, no. Kubeconfig files are parsed by kubectl and the code is quite elaborate. At this time I don't think
adding it to ContainerSSH is wise.

### Why does the `kuberun` backend have so many things it doesn't support?

The `kuberun` backend was written by reverse engineering `kubectl`. Unfortunately the Kubernetes API is documented very
poorly and is quirky in some places. Kubernetes is a very complex and fast moving beast so things like API
documentation, a proper SDK and other niceties that make a developers life easy are not something that's currently 
available.
