<h1>The <code>dockerrun</code> backend</h1>

The `dockerrun` backend launches a container using the Docker API

## Providing certificates

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

## Changing the container image

The container image depends on the backend you are using. For `dockerrun` you can change the image in the config
file:

```yaml
dockerrun:
  config:
    container:
      image: your/image
``` 

## Detailed configuration

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
