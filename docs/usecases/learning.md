---
title: Learning environments
---

Providing console access to your students can be a hurdle. No matter if you want to provide access to a Linux environment, databases, or something else, you will have to create users, make sure they have everything they need, and clean up when they are done.

Containers can provide an easier solution: launch a specific environment for a student and simply remove the container when they are done. However, manually provisioning containers can still be tedious and continuously running a large number of containers can be resource-intensive.

ContainerSSH provides a vital role here: it can dynamically launch containers as needed. When users connect via SSH, ContainerSSH reaches out to [your authentication server](../authserver.md) to verify user credentials and then contacts [your configuration server](../configserver.md) to fetch the customized container configuration for your user. When your user disconnects, ContainerSSH removes their container and leaves no trace behind.

[Get started »](../quickstart.md){: .md-button}