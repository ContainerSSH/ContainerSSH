# This Dockerfile is built as a quick example to deploy ContainerSSH.
# It is intended to be run on a local Docker setup for testing purposes.
FROM containerssh/containerssh:0.3.1
# We are changing the user back to 0 so we can access the Docker socket.
# We don't recommend using this setting in production. Please take a look
# at https://containerssh.io/reference/dockerrun/ for hardened settings.
USER 0
