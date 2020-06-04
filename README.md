# An SSH server that launches containers

This is a Proof of Concept SSH server written in Go that sends any shell directly into a Docker container instead
of launching it on a local machine. It uses an HTTP microservice as an authentication endpoint for SSH connections.

## Building

The project can be build using `go build`.

## Running

In order to run the application you will need to generate at least one SSH host key, for example with
`ssh-keygen -t rsa`.

You can then run the application by specifying either `--auth-password` or `--auth-pubkey`.

```
./containerssh --hstkey-rsa ~/.ssh/id_rsa --auth-password
```

For additional configuration options see `-h`.

## Implementing an authentication server

The authentication endpoint needs to respond to two URLs: `/password` and `/pubkey`.

The password endpoint will receive a request in this format:

```json
{
    "user": "username",
    "remoteAddress": "127.0.0.1:1234",
    "sessionIdBase64": "A base64 SSH session ID",
    "passwordBase64": "Base 64 password"
}
```

The public key endpoint will receive requests in this format:

```json
{
    "user": "username",
    "remoteAddress": "127.0.0.1:1234",
    "sessionIdBase64": "A base64 SSH session ID",
    "publicKeyBase64": "Base 64 public key in SSH wire format"
}
```

Both endpoints need to respond with the following JSON:

```json
{
  "success": true
}
```