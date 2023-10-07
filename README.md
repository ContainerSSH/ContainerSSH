[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">An SSH Server that Launches Containers in Kubernetes and Docker</h1>

[![Documentation: available](https://img.shields.io/badge/documentation-available-green?style=for-the-badge)](https://containerssh.io/)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/containerssh/containerssh/goreleaser?style=for-the-badge)](https://github.com/containerssh/containerssh/actions)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/containerssh/containerssh?sort=semver&style=for-the-badge)](https://github.com/containerssh/containerssh/releases)
[![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/containerssh/containerssh?style=for-the-badge)](http://hub.docker.com/r/containerssh/containerssh)
[![Go Report Card](https://goreportcard.com/badge/github.com/containerssh/containerssh?style=for-the-badge)](https://goreportcard.com/report/github.com/containerssh/containerssh)
[![LGTM Alerts](https://img.shields.io/lgtm/alerts/github/ContainerSSH/ContainerSSH?style=for-the-badge)](https://lgtm.com/projects/g/ContainerSSH/ContainerSSH/)
[![License: Apache 2.0](https://img.shields.io/github/license/ContainerSSH/ContainerSSH?style=for-the-badge)](LICENSE.md)

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FContainerSSH%2FContainerSSH.svg?type=shield&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2FContainerSSH%2FContainerSSH?ref=badge_shield&issueType=license)

## ContainerSSH in One Minute

In a hurry? This one-minute video explains everything you need to know about ContainerSSH.

[![An image with a YouTube play button on it.](https://containerssh.io/images/containerssh-intro-preview.png)](https://youtu.be/Cs9OrnPi2IM)

## Need help?

[Join the #containerssh Slack channel on the CNCF Slack »](https://communityinviter.com/apps/cloud-native/cncf)

## Use cases

### Build a lab

Building a lab environment can be time-consuming. ContainerSSH solves this by providing dynamic SSH access with APIs, automatic cleanup on logout using ephemeral containers, and persistent volumes for storing data. **Perfect for vendor and student labs.**

[Read more »](https://containerssh.io/usecases/lab/)

### Debug a production system

Provide **production access to your developers**, give them their usual tools while logging all changes. Authorize their access and create short-lived credentials for the database using simple webhooks. Clean up the environment on disconnect.

[Read more »](https://containerssh.io/usecases/debugging/)

### Run a honeypot

Study SSH attack patterns up close. Drop attackers safely into network-isolated containers or even virtual machines, and **capture their every move** using the audit logging ContainerSSH provides. The built-in S3 upload ensures you don't lose your data.

[Read more »](https://containerssh.io/usecases/honeypots/)

## How does it work?

![](https://containerssh.io/images/architecture.svg)

1. The user opens an SSH connection to ContainerSSH.
2. ContainerSSH calls the authentication server with the users username and password/pubkey to check if its valid.
3. ContainerSSH calls the config server to obtain backend location and configuration (if configured)
4. ContainerSSH calls the container backend to launch the container with the
   specified configuration. All input from the user is sent directly to the backend, output from the container is sent
   to the user.

[▶️ Watch as video »](https://youtu.be/Cs9OrnPi2IM) | [🚀 Get started »](https://containerssh.io/quickstart/)

## Demo

![](https://containerssh.io/images/ssh-in-action.gif)

[🚀 Get started »](https://containerssh.io/quickstart/)

## Contributing

If you would like to contribute, please check out our [Code of Conduct](https://github.com/ContainerSSH/community/blob/main/CODE_OF_CONDUCT.md) as well as our [contribution documentation](https://containerssh.io/development/).
