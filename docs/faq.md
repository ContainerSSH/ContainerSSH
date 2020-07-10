<h1>FAQ</h1>

## Is ContainerSSH secure?

ContainerSSH depends on a number of libraries to achieve what it does. A security hole in any of the critical ones
could mean a compromise of your container environment, especially if you are using the `dockerrun` backend. (Docker
has no access control so a compromise means your whole host is compromised.)

## Is ContainerSSH production-ready?

No. ContainerSSH is very early in its development and has not undergone extensive testing yet. You should be careful
before deploying it into production.

## Does ContainerSSH delete containers after it is done?

ContainerSSH does its best to delete containers it creates. However, at this time there is no cleanup mechanism in case
it crashes.

## Do I need to run ContainerSSH as root?

No! In fact, you shouldn't! ContainerSSH is perfectly fine running as non-root as long as it has access to Kubernetes
or Docker. (Granted, access to the Docker socket means it could easily launch a root process on the host.)

## Does ContainerSSH support SFTP?

Yes, but your container image must contain an SFTP server binary and your config.yaml or config server must contain the
correct path for the server.

## Does ContainerSSH support SCP?

Not at this time.

## Does ContainerSSH support TCP port forwarding?

Not at this time. TCP port forwarding is done outside of a channel. At this time ContainerSSH launches one container
per SSH channel, so forwarding TCP ports would mean a complete overhaul of the entire architecture. In essence the
architecture would be changed to launch one container per session, not per channel, and use `exec` to launch a shell or
SFTP server for the channel.

However, as you might imagine that's a bit of a larger change and will need quite a bit of work.

## Does ContainerSSH support SSH agent forwarding?

Not at this time. SSH agent forwarding would require a separate binary agent within the container to proxy data.
This is similar to how TCP port forwarding works, except that the authentication agent requests are sent on a
per-channel basis. Additionally SSH agent forwarding is not documented well, it is proprietary to OpenSSH.
(The request type is `auth-agent-req@openssh.com`.)

## Does ContainerSSH support X11 forwarding?

Not at this time. X11 is sent over separate channels and would most probably need the overhaul that TCP port forwarding
requires. As X11 forwarding isn't use much any more it is unlikely that ContainerSSH will ever support it.

## Does ContainerSSH support forwarding signals?

Partially. The `dockerrun` backend supports it, the `kuberun` backend doesn't because Kubernetes itself doesn't.

## Does ContainerSSH support window resizing?

Yes.

## Does ContainerSSH support environment variable passing?

Yes.

## Does ContainerSSH support returning the exit status?

Partially. The `dockerrun` backend supports it, the `kuberun` backend &ldquo;does its best&rdquo; but has some
edge cases when the connection closes before the exit status can be obtained.

## Can ContainerSSH run exec into existing containers?

Not at this time. The architecture needs to solidify before such a feature is implemented.

## Can ContainerSSH deploy additional services, such as sidecar containers, etc?

ContainerSSH supports the entire Kubernetes pod specification so you can launch as many containers as you want in a
single pod. The Docker backend, however, does not support sidecar containers.

## Can I add metadata to my pods with the `kuberun` backend?

Not at this time. You may want to open up a feature request and detail your use case.

## Why is the `kuberun` backend so slow?

Kubernetes is built for scale. That means there are some tradeoffs in terms of responsiveness. This is not something
ContainerSSH can do anything about, it just takes a bit to launch a pod. You may want to fine-tune your Kubernetes 
cluster for responsiveness.

## Why is there no initial prompt with the `kuberun` backend?

This is a [known bug](https://github.com/janoszen/containerssh/issues/12). Unfortunately the `kuberun` backend was 
built by reverse engineering kubectl as there is no documentation whatsoever on how the attach functionality works on
pods. If you are good with Go you might want to help out here.

## Can I use my normal kubeconfig files?

Unfortunately, no. Kubeconfig files are parsed by kubectl and the code is quite elaborate. At this time I don't think
adding it to ContainerSSH is wise.

## Why does the `kuberun` backend have so many things it doesn't support?

The `kuberun` backend was written by reverse engineering `kubectl`. Unfortunately the Kubernetes API is documented very
poorly and is quirky in some places. Kubernetes is a very complex and fast moving beast so things like API
documentation, a proper SDK and other niceties that make a developers life easy are not something that's currently 
available.
