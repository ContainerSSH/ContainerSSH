<h1>Quick start</h1>

This is a quick start guide to get a test server up and running in less than 5 minutes with [docker-compose](https://docs.docker.com/compose/).

## Step 1: Set up a Dockerized environment

To run this quick start please make sure you have a working [Docker environment](https://docs.docker.com/get-docker/)
and a working [docker-compose](https://docs.docker.com/compose/).

## Step 2: download the sample files

Please download the contents of the [example directory](https://github.com/janoszen/containerssh/tree/stable/example)
from the source code repository.

## Step 3: Launch ContainerSSH

In the downloaded directory run `docker-compose build` and then `docker-compose up`.

## Step 4: Logging in

Run `ssh foo@localhost -p 2222` on the same machine. You should be able to log in with any password.

Alternatively you can also try the user `busybox` to land in a Busybox container.

## Step 5: Making it productive

The authentication and configuration server included in the example is a dummy server and lets any password in. To
actually use ContainerSSH you will have to write [your own authentication server](authserver.md) and you may want to
write your own [configuration server too](configserver.md).
