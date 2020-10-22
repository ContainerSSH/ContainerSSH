---
template: overrides/home.html
title: "ContainerSSH: An SSH server that launches containers"
---

<h1>ContainerSSH: launch containers on-demand</h1>

<div class="grid">
<div class="grid__box">
<h2>Offering SSH in a web hosting service?</h2>
<p>ContainerSSH lets you dynamically create and destroy containers when your users connect. Authenticate against your existing user database and mount directories based on your existing permission matrix.</p>
<p><a href="usecases/webhosting/" class="md-button">Read more »</a></p>
</div>
<div class="grid__box">
<h2>Looking for a Linux learning environment?</h2>
<p>With ContainerSSH you can launch Linux-based containers on-demand when your students connect. You can supply your own container image and mount folders with learning and testing material as needed.</p>
<p><a href="usecases/learning/" class="md-button">Read more »</a></p>
</div>
<div class="grid__box">
<h2>Building a honeypot?</h2>
<p>With the dynamic authentication server of ContainerSSH you can capture usernames and passwords, and you container environment can log commands that are executed.</p>
<p><a href="usecases/honeypots/" class="md-button">Read more »</a></p>
</div>
<div class="grid__box">
<h2>Building a high security environment?</h2>
<p>ContainerSSH is being used to provide dynamic console access to an environment with sensitive credentials. Use the authentication and configuration server to dynamically provision credentials in conjunction with secret management systems such as Hashicorp Vault.</p>
<p><a href="usecases/security/" class="md-button">Read more »</a></p>
</div>
</div>

## How does it work?

![](https://projects.pasztor.at/containerssh/images/ssh-in-action.gif)

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
   
[Get started »](quickstart.md){: .md-button}