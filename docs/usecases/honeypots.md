---
title: Honeypots
---

When left undefended, SSH can be a large attack surface towards the internet. If you leave an SSH server open to the Internet, bots will try to brute force their way in within minutes. Why not build a honeypot?

Honeypots can provide valuable early warning: log the IP addresses of connection attempts and dynamically firewall them. Collect credentials attackers are trying to use and match them against your user database to root out weak passwords. Collect logs of what attackers are doing in a containerized environment.

ContainerSSH can do all that. When a user connects, ContainerSSH reaches out to your [authentication server](../authserver.md) where you can log IP addresses and credentials.

If you allow attackers to connect, ContainerSSH fetches a dynamic container configuration from your [configuration server](../configserver.md). You can specify what environment and on which Docker or Kubernetes setup to run your honeypot. Restrict attackers to a set amount of resources or a networkless environment.

[Get started »](../quickstart.md){: .md-button}