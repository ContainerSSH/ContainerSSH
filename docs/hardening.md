<h1>Hardening ContainerSSH</h1>

ContainerSSH is built to secure its inner workings as much as possible. You can take several steps to secure it further.

## Running ContainerSSH

The default [ContainerSSH image](https://hub.docker.com/r/janoszen/containerssh) runs as a non-root user by default
and exposes itself on port 2222. If you decide to build your own installation make sure ContainerSSH does not run
as root as it is not required.

## Secure your Docker/Kubernetes

Depending on which backend you are using you have to take different steps to secure it.

When using Docker ContainerSSH will need access to the Docker socket. This undeniably means that ContainerSSH will
be able to launch root processes on the host machine. You may want to look into running Docker in
[rootless mode](https://docs.docker.com/engine/security/rootless/) or switching to [Podman](https://podman.io/)

When running Kubernetes it is strongly advised that you deploy a pod security policy and a network policy. You should
also make sure that ContainerSSH uses a restricted service account that can only access its own namespace.

## Securing your auth server

Your authentication server contains all your secrets and is therefore a prime target. ContainerSSH delegates any and
all access checking to the authentication server so you should make sure it prevents brute force attacks.

Furthermore, you should make sure that the authentication server cannot be accessed from anywhere else. You can do this
using firewalls, or alternatively you can configure ContainerSSH to use client certificates to authenticate itself:

```yaml
auth:
    url: http://127.0.0.1:8080
    cacert: "insert your expected CA certificate in PEM format here"
    timeout: 2s
    cert: "insert your client certificate in PEM format here"
    key: "insert your client key in PEM format here"
```

## Securing your config server

Similar to your authentication server you can also secure the config server in a similar manner:

```yaml
configserver:
    timeout: 2s
    url: http://127.0.0.1:8080/config
    cacert: "insert your expected CA certificate in PEM format here"
    cert: "insert your client certificate in PEM format here"
    key: "insert your client key in PEM format here"
```

## Disabling command execution {{ since("0.2.1") }}

You can disable the execution of custom SSH commands through the configuration:


```yaml
dockerrun:
    config:
        disableCommand: true
```

```yaml
kuberun:
    pod:
        disableCommand: true
```

!!! note
    Enabling command execution also disables SFTP integration.