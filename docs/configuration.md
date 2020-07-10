<h1>Configuring ContainerSSH</h1>

Before you can run ContainerSSH you will need to create a configuration file. The minimal configuration file looks like
this:

```yaml
ssh:
  hostkeys:
    # Generate a host key with ssh-keygen -t rsa
    - /path/to/your/host/key
auth:
  # See auth server below
  url: http://your-auth-server/
  password: true # Perform password authentication
  pubkey: false # Perform public key authentication
```

The config file must end in `.yml`, `.yaml`, or `.json`. You can dump the entire configuration file using
`./containerssh --dump-config`

!!! note
    Parts of the configuration can be provided dynamically based on the username using a [configserver](configserver.md)

!!! note
    In order to actually use ContainerSSH you will also need to provide [a backend configuration](backends.md) wither
    via this file or via the [configserver](configserver.md). 
