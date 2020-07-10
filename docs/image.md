<h1>Building a container image for ContainerSSH</h1>

ContainerSSH has no requirements as to the container image you are running apart from the fact that they need to be 
Linux containers.

If you wish to use SFTP you have to add an SFTP server (`apt install openssh-sftp-server` on Ubuntu) to the container
image and configure the path of the SFTP server correctly in your config.yaml. The sample image
`janoszen/containerssh-image` contains an SFTP server.
