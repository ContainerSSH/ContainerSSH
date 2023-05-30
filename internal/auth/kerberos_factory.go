package auth

import (
	"os"

	krb5cfg "github.com/containerssh/gokrb5/v8/config"
	"github.com/containerssh/gokrb5/v8/iana/nametype"
	"github.com/containerssh/gokrb5/v8/keytab"
	"github.com/containerssh/gokrb5/v8/types"
	"go.containerssh.io/libcontainerssh/config"
	"go.containerssh.io/libcontainerssh/internal/metrics"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
)

func NewKerberosClient(
	authType AuthenticationType,
	cfg config.AuthKerberosClientConfig,
	logger log.Logger,
	_ metrics.Collector,
) (KerberosClient, error) {
	if err := cfg.Validate(); err != nil {
		return nil, message.Wrap(
			err,
			message.EAuthConfigError,
			"Kerberos configuration failed to validate",
		)
	}

	kt, err := keytab.Load(cfg.Keytab)
	if err != nil {
		return nil, message.Wrap(
			err,
			message.EAuthConfigError,
			"Failed to load kerberos keytab from %s",
			cfg.Keytab,
		)
	}

	var conf *krb5cfg.Config
	if authType == AuthenticationTypePassword {
		conf, err = krb5cfg.Load(cfg.ConfigPath)
		if err != nil {
			return nil, message.Wrap(
				err,
				message.EAuthConfigError,
				"Failed to load kerberos configuration file from %s",
				cfg.ConfigPath,
			)
		}
	}

	var acceptor *types.PrincipalName
	if cfg.Acceptor == "any" {
		acceptor = nil
	} else if cfg.Acceptor == "host" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, message.Wrap(
				err,
				message.EAuthConfigError,
				"Failed to get hostname from OS",
			)
		}
		acceptor = &types.PrincipalName{
			NameType:   nametype.KRB_NT_PRINCIPAL,
			NameString: []string{"host", hostname},
		}
	} else {
		a := types.NewPrincipalName(nametype.KRB_NT_PRINCIPAL, cfg.Acceptor)
		acceptor = &a
	}

	return &kerberosAuthClient{
		authType:   authType,
		logger:     logger,
		config:     cfg,
		keytab:     kt,
		acceptor:   acceptor,
		clientConf: conf,
	}, nil
}
