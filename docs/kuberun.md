<h1>The <code>kuberun</code> backend</h1>

The kuberun backend runs a pod in a Kubernetes cluster and attaches to a container there.

## Running outside of Kubernetes

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

## Running inside a Kubernetes cluster

When you run inside of a Kubernetes cluster you can use the service account token:

```yaml
kuberun:
    connection:
        certFile: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
```

## Changing the container image

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

## Detailed configuration

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
