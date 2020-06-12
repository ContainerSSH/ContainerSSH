# ContainerSSH quick start

This is a quick start example for [ContainerSSH](https://github.com/janoszen/containerssh), an SSH server that launches
Docker containers and Kubernetes pods.

This example utilizes [docker-compose](https://docs.docker.com/compose/) in a Docker environment to launch ContainerSSH
together with a dummy authentication server.

Launch it using `docker-compose up` and then SSH into `localhost` port 2222. You can use the following users:

- `foo` with any password to get an Ubuntu container with SFTP support
- `busybox` with any password to get a Busybox container