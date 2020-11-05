[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.github.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">An SSH Server that Launches Containers in Kubernetes and Docker</h1>

[![Documentation: available](https://img.shields.io/badge/documentation-available-green?style=for-the-badge)](https://containerssh.github.io/)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/containerssh/containerssh/goreleaser?style=for-the-badge)](https://github.com/containerssh/containerssh/actions)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/containerssh/containerssh?sort=semver&style=for-the-badge)](https://github.com/containerssh/containerssh/releases)
[![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/containerssh/containerssh?style=for-the-badge)](http://hub.docker.com/r/containerssh/containerssh)
[![Go Report Card](https://goreportcard.com/badge/github.com/containerssh/containerssh?style=for-the-badge)](https://goreportcard.com/report/github.com/containerssh/containerssh)
[![LGTM Alerts](https://img.shields.io/lgtm/alerts/github/ContainerSSH/ContainerSSH?style=for-the-badge)](https://lgtm.com/projects/g/ContainerSSH/ContainerSSH/)
[![GitHub](https://img.shields.io/github/license/containerssh/containerssh?style=for-the-badge)](LICENSE.md)

## Offering SSH in a web hosting service?

ContainerSSH lets you dynamically create and destroy containers when your users connect. Authenticate against your existing user database and mount directories based on your existing permission matrix.

[Read more »](https://containerssh.github.io/usecases/webhosting/)

## Looking for a Linux learning environment?

With ContainerSSH you can launch Linux-based containers on demand when your students connect. You can supply your own container image and mount folders with learning and testing material as needed.</p>

[Read more »](https://containerssh.github.io/usecases/learning/)

## Building a honeypot?

With the dynamic authentication server of ContainerSSH you can capture usernames and passwords, and you container environment can log commands that are executed.

[Read more »](https://containerssh.github.io/usecases/honeypots/)

## Building a high security environment?

ContainerSSH is being used to provide dynamic console access to an environment with sensitive credentials. Use the authentication and configuration server to dynamically provision credentials in conjunction with secret management systems such as Hashicorp Vault.

[Read more »](https://containerssh.github.io/usecases/security/)

## How does it work?

![](https://containerssh.github.io/images/ssh-in-action.gif)

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
   
[Get started »](https://containerssh.github.io/quickstart/)
