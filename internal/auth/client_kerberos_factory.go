package auth

import (
	"fmt"
	"os"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/log"

	"github.com/containerssh/gokrb5/v8/keytab"
	krb5cfg "github.com/containerssh/gokrb5/v8/config"
	"github.com/containerssh/gokrb5/v8/types"
	"github.com/containerssh/gokrb5/v8/iana/nametype"
)

func NewKerberosClient(
	cfg config.AuthConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (Client, error) {
	if cfg.Method != config.AuthMethodKerberos {
		return nil, fmt.Errorf("authentication is not set to kerberos")
	}

	if err := cfg.Validate(); err != nil {
		return nil, message.Wrap(
			err,
			message.EAuthConfigError,
			"Kerberos configuration failed to validate",
		)
	}

	kt, err := keytab.Load(cfg.Kerberos.Keytab)
	if err != nil {
		return nil, message.Wrap(
			err,
			message.EAuthConfigError,
			"Failed to load kerberos keytab from %s",
			cfg.Kerberos.Keytab,
		)
	}

	var conf *krb5cfg.Config
	if cfg.Kerberos.AllowPassword {
		conf, err = krb5cfg.Load(cfg.Kerberos.ConfigPath)
		if err != nil {
			return nil, message.Wrap(
				err,
				message.EAuthConfigError,
				"Failed to load kerberos configuration file from %s",
				cfg.Kerberos.ConfigPath,
			)
		}
	}

	var acceptor *types.PrincipalName
	if cfg.Kerberos.Acceptor == "any" {
		acceptor = nil
	} else if cfg.Kerberos.Acceptor == "host" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, message.Wrap(
				err,
				message.EAuthConfigError,
				"Failed to get hostname from OS",
			)
		}
		acceptor = &types.PrincipalName{
			NameType: nametype.KRB_NT_PRINCIPAL,
			NameString: []string{"host", hostname},
		}
	} else {
		a := types.NewPrincipalName(nametype.KRB_NT_PRINCIPAL, cfg.Kerberos.Acceptor)
		acceptor = &a
	}

	return &kerberosAuthClient{
		logger: logger,
		config: cfg.Kerberos,
		keytab: kt,
		acceptor: acceptor,
		clientConf: conf,
	}, nil
}
