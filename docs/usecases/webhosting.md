---
title: ContainerSSH for web hosting
---

Providing SSH access in a web hosting environment is tricky. Users may run unexpected scripts that consume lots of resources. They may have permission issues if they are not able to SSH with the same user as the webserver, which in turn presents security issues.

Containers present a good solution for this problem: you can run a container as the same user as the web server, but keep them in isolation from the actual production environment. You can use NFS mounts to isolate them from the production servers. You can even mount folders based on an advanced permission matrix.

However, running an SSH server per user is very cost-intensive in an industry where individual customers don't pay much. That's where ContainerSSH fills an important role: when users connect via SSH, ContainerSSH reaches out to [your authentication server](../authserver.md) to verify user credentials and then contacts [your configuration server](../configserver.md) to fetch the customized container configuration for your user. When your user disconnects, ContainerSSH removes their container and leaves no trace behind.

ContainerSSH also supports SFTP, which provides secure file transfers. It can replace the old and arguably broken FTP, so you no longer have to worry about that either.

If you are running multiple servers, you can even provide dynamic Docker connect strings and credentials to connect the server where the user is located.

[Get started »](../quickstart.md){: .md-button}