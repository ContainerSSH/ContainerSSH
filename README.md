# An SSH server that launches containers

This is a Proof of Concept SSH server written in Go that sends any shell directly into a Docker container instead
of launching it on a local machine. It uses an HTTP microservice as an authentication endpoint for SSH connections.

## Building

The project can be built using `make build` or `make build-docker`.

## Configuring

Before you can run containerssh you will need to create a configuration file. The minimal configuration file looks like
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

Note that the config file must end in `.yml`, `.yaml`, or `.json`. You can dump the entire configuration file using
`./containerssh --dump-config`

## Running

You can run the containerssh server using the following command line:

```
./containerssh --config your/config-file.yaml
```

## Implementing an authentication server

Containerssh does not know your users and their passwords. Therefore it calls out to a microservice that you have to
provide so it can verify the users, passwords and SSH keys. You will have to provide the microservice URL in the
configuration.

For password authentication containerssh will call out to `http://your-auth-server/password` with the following request
body. The password is base64 encoded to transfer special characters properly.

```json
{
    "user": "username",
    "remoteAddress": "127.0.0.1:1234",
    "sessionId": "A base64 SSH session ID",
    "passwordBase64": "Base 64 password"
}
```

The public key auth containerssh will call out to `http://your-auth-server/pubkey` in the following format.

```json
{
    "user": "username",
    "remoteAddress": "127.0.0.1:1234",
    "sessionId": "A base64 SSH session ID",
    "publicKeyBase64": "Base 64 public key in SSH wire format"
}
```

The public key is provided in the SSH wire format in base64 encoding.

Both endpoints need to respond with the following JSON:

```json
{
  "success": true
}
```

## Changing the container image

The container image depends on the backend you are using. For `dockerrun` you can change the image in the config
file:

```yaml
dockerrun:
  config:
    container:
      image: your/image
``` 

## Providing certificates for `dockerrun`

Docker sockets allow connections over TCP with TLS encryption. You can provide these TLS certificates embedded in the
YAML file in PEM format:

```yaml
dockerrun:
  host: tcp://127.0.0.1:2376
  cacert: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  cert: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  key: |
    -----BEGIN RSA PRIVATE KEY-----
    -----END RSA PRIVATE KEY-----
```

## Configuration server

You have the option to dynamically change the configuration based on the username by providing a config server URL:

```yaml
configserver:
  url: http://your-config-server-url/
```

The config server will receive a request in following format:

```json
{
  "username":"ssh username",
  "sessionId": "ssh session ID"
}
```

Your application will have to respond in the following format:

```json
{
  "config": {
    // Provide a partial configuration here 
  }
}
```

You can view the full configuration structure in YAML format by running `./containerssh --dump-config`. Note that your
config server must respond in JSON format.

Some configuration values cannot be overridden from the config server. These are the ones that get used before the
connection is established, but the Docker provider also overrides a few, such as `attachstdio`.
