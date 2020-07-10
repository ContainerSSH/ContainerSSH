<h1>Quick start</h1>

This is a quick start guide to get a test server up and running in less than 5 minutes with [docker-compose](https://docs.docker.com/compose/).

To run it grab all files from the [example directory](https://github.com/janoszen/containerssh/tree/stable/example) and
run `docker-compose build` followed by `docker-compose up` in that directory. This will run the SSH server on your local
machine on port 2222. You can log in with any password using the user "foo" to get an Ubuntu image and "busybox" to get
a Busybox image.

Note that this launches a dummy authentication server which is (obviously) not suited for your production environment.
To actually use ContainerSSH you will have to write [your own authentication server](authserver.md).
