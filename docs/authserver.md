<h1>Implementing an authentication server</h1>

!!! note
    We have an [OpenAPI document](/api/authconfig) available for the authentication and configuration server. You can
    check the exact values available there, or use the OpenAPI document to generate parts of your server code.

ContainerSSH does not know your users and their passwords. Therefore, it calls out to a microservice that you have to
provide so it can verify the users, passwords and SSH keys. You will have to provide the microservice URL in the
configuration.

For password authentication ContainerSSH will call out to `http://your-auth-server/password` with the following request
body. The password is base64 encoded to transfer special characters properly.

```json
{
    "username": "username",
    "remoteAddress": "127.0.0.1:1234",
    "sessionId": "A base64 SSH session ID",
    "passwordBase64": "Base 64 password"
}
```

> **Note:** Earlier versions of ContainerSSH used the `user` field instead of `username`. While the `user` field still
> exists it is considered deprecated and will be removed in a future version.

The public key auth ContainerSSH will call out to `http://your-auth-server/pubkey` in the following format.

```json
{
    "username": "username",
    "remoteAddress": "127.0.0.1:1234",
    "sessionId": "A base64 SSH session ID",
    "publicKeyBase64": "Base 64 public key in SSH wire format"
}
```

> **Note:** Earlier versions of ContainerSSH used the `user` field instead of `username`. While the `user` field still
> exists it is considered deprecated and will be removed in a future version.

The public key is provided in the SSH wire format in base64 encoding.

Both endpoints need to respond with the following JSON:

```json
{
  "success": true
}
```

> **Tip** You can find the source code for a test authentication and configuration server written in Go
> [in the code repository](https://github.com/janoszen/containerssh/blob/stable/cmd/containerssh-testauthconfigserver/main.go)
