<h1>Writing a configuration server</h1>

ContainerSSH has the ability to configure the backend and the launched container dynamically based on the username
and/or IP address. To do this ContainerSSH calls out to a configuration server if configured.

You have the option to dynamically change the configuration based on the username by providing a config server URL:

```yaml
configserver:
  url: http://your-config-server-url/
```

Once you have this configured you can launch a HTTP server that returns a configuration fragment as described below
on that address.

!!! note
    We have an [OpenAPI document](../api/authconfig) available for the authentication and configuration server. You can
    check the exact values available there, or use the OpenAPI document to generate parts of your server code.

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
connection is established, but the backends also override a few, such as `attachstdio`.