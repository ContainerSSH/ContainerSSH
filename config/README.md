# ContainerSSH configuration

This package contains the configuration structure of ContainerSSH. The root
configuration object is `config.AppConfig`. You can use this library to
generate a valid ContainerSSH configuration file.

Each configuration structure also contains a `Validate()` function which
can be used to check if the configuration is valid. (Note, this will do a
full validation. Validating a partial configuration is not supported.)