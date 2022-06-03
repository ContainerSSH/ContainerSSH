package auth

import (
	"fmt"
	"net"

	"github.com/containerssh/gokrb5/v8/asn1tools"
	"github.com/containerssh/gokrb5/v8/client"
	krbconf "github.com/containerssh/gokrb5/v8/config"
	"github.com/containerssh/gokrb5/v8/credentials"
	"github.com/containerssh/gokrb5/v8/gssapi"
	"github.com/containerssh/gokrb5/v8/iana/keyusage"
	"github.com/containerssh/gokrb5/v8/keytab"
	krbmsg "github.com/containerssh/gokrb5/v8/messages"
	krb5svc "github.com/containerssh/gokrb5/v8/service"
	"github.com/containerssh/gokrb5/v8/spnego"
	"github.com/containerssh/gokrb5/v8/types"
    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/ssh"
    internalSsh "go.containerssh.io/libcontainerssh/internal/ssh"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
    "go.containerssh.io/libcontainerssh/metadata"
	"gopkg.in/jcmturner/goidentity.v3"
)

type kerberosAuthContext struct {
	client *kerberosAuthClient

	connectionId string
	remoteAddr   net.IP

	key               types.EncryptionKey
	principalUsername string
	loginUsername     string
	credentials       []byte

	meta    metadata.ConnectionAuthenticatedMetadata
	success bool
	err     error
}

func (k kerberosAuthContext) AuthenticatedUsername() string {
	return k.principalUsername
}

type kerberosAuthClient struct {
	logger     log.Logger
	config     config.AuthKerberosClientConfig
	keytab     *keytab.Keytab
	clientConf *krbconf.Config
	acceptor   *types.PrincipalName
	authType   AuthenticationType
}

func (k kerberosAuthContext) Success() bool {
	return k.success
}

func (k kerberosAuthContext) Error() error {
	if !k.success && k.err == nil {
		k.err = fmt.Errorf("an unknown error happened during kerberos authentication")
	}
	return k.err
}

func (k kerberosAuthContext) Metadata() metadata.ConnectionAuthenticatedMetadata {
	if k.client == nil {
		return k.meta
	}
	meta := k.meta
	if k.client.config.CredentialCachePath != "" && k.credentials != nil {
		path := k.client.config.CredentialCachePath
		meta.GetFiles()[path] = metadata.BinaryValue{
			Value:     k.credentials,
			Sensitive: true,
		}
		meta.GetEnvironment()["KRB5CCNAME"] = metadata.Value{
			Value: "FILE:" + k.client.config.CredentialCachePath,
		}
	}
	return meta
}

func (k kerberosAuthContext) OnDisconnect() {

}

func (c *kerberosAuthClient) Password(
	meta metadata.ConnectionAuthPendingMetadata,
	password []byte,
) AuthenticationContext {
	if c.authType != AuthenticationTypePassword && c.authType != AuthenticationTypeAll {
		return &kerberosAuthContext{
			meta:    meta.AuthFailed(),
			success: false,
			err:     fmt.Errorf("authentication client not configured for password authentication"),
		}
	}

	cl := client.NewWithPassword(
		meta.Username,
		c.clientConf.LibDefaults.DefaultRealm,
		string(password),
		c.clientConf,
		client.DisablePAFXFAST(true), // PAFXFAST breaks Active-Directory, see https://github.com/jcmturner/gokrb5/blob/master/USAGE.md#active-directory-kdc-and-fast-negotiation
	)

	err := cl.Login()
	if err != nil {
		return kerberosAuthContext{
			client:  c,
			success: false,
			err:     err,
		}
	}

	ccache, err := cl.GetCCache()
	if err != nil {
		return kerberosAuthContext{
			client:  c,
			success: false,
			err:     err,
		}
	}

	ccacheMar, err := ccache.Marshal()
	if err != nil {
		return kerberosAuthContext{
			client:  c,
			success: false,
			err:     err,
		}
	}

	ctx := kerberosAuthContext{
		client:            c,
		principalUsername: meta.Username,
		loginUsername:     meta.Username,
		credentials:       ccacheMar,
		connectionId:      meta.ConnectionID,
		remoteAddr:        meta.RemoteAddress.IP,
		success:           true,
		err:               nil,
	}

	authMeta, err := ctx.AllowLogin(meta.Username, meta)
	if err != nil {
		return kerberosAuthContext{
			client:  c,
			success: false,
			err:     err,
		}
	}

	return kerberosPasswordAuthContext{
		ctx,
		authMeta,
	}
}

type kerberosPasswordAuthContext struct {
	kerberosAuthContext

	meta metadata.ConnectionAuthenticatedMetadata
}

func (c *kerberosAuthClient) GSSAPI(meta metadata.ConnectionMetadata) GSSAPIServer {
	if c.authType != AuthenticationTypeGSSAPI && c.authType != AuthenticationTypeAll {
		return &kerberosAuthContext{
			connectionId: meta.ConnectionID,
			remoteAddr:   meta.RemoteAddress.IP,
			success:      false,
			err:          fmt.Errorf("authentication client not configured for GSSAPI authentication"),
		}
	}
	return &kerberosAuthContext{
		client:       c,
		connectionId: meta.ConnectionID,
		remoteAddr:   meta.RemoteAddress.IP,
	}
}

func (k *kerberosAuthContext) AcceptSecContext(token []byte) (outputToken []byte, srcName string, needContinue bool, err error) {
	var st spnego.KRB5Token
	err = st.Unmarshal(token)
	if err != nil {
		return nil, "", false, message.Wrap(
			err,
			message.EAuthKerberosVerificationFailed,
			"Failed to unmarshal the intial KRB token",
		)
	}
	st.Settings = krb5svc.NewSettings(
		k.client.keytab,
		krb5svc.MaxClockSkew(k.client.config.ClockSkew),
	)
	verified, _ := st.Verify()

	if verified {
		ctx := st.Context()
		id := ctx.Value(spnego.CtxCredentials).(goidentity.Identity)
		k.principalUsername = id.UserName()

		a := st.APReq

		hostAddr := types.HostAddressFromNetIP(k.remoteAddr)

		ok, err := a.Verify(k.client.keytab, k.client.config.ClockSkew, hostAddr, k.client.acceptor)
		if err != nil {
			return nil, "", false, message.Wrap(
				err,
				message.EAuthKerberosVerificationFailed,
				"Failed to verify the AP_REQ packet (is the acceptor correct?)",
			)
		}
		if !ok {
			return nil, "", false, message.Wrap(
				err,
				message.EAuthKerberosVerificationFailed,
				"Couldn't verify AP_REQ packet",
			)
		}

		k.key = a.Authenticator.SubKey

		ticket := a.Ticket
		rep, err := krbmsg.NewAPRep(ticket, a.Authenticator)
		if err != nil {
			return nil, "", false, message.Wrap(
				err,
				message.EAuthKerberosVerificationFailed,
				"Failed to generate an AP_REP packet",
			)
		}

		repToken := spnego.NewKRB5TokenAPREP()
		repToken.APRep = rep
		mar2, err := repToken.Marshal()
		asn1tools.AddASNAppTag(mar2, 060)
		if err != nil {
			return nil, "", false, message.Wrap(
				err,
				message.EAuthKerberosVerificationFailed,
				"Failed to marshal the AP_REP packet",
			)
		}

		authCred, err := a.Authenticator.GetCredDelegation()
		if err != nil {
			k.client.logger.Info(message.Wrap(
				err,
				message.EAuthKerberosVerificationFailed,
				"Failed to unmarshal the Authenticator delegation packet",
			))
			// Accept but no cred delegation
			return mar2, k.principalUsername, false, nil
		}

		if authCred != nil && authCred.HasDelegation() {
			var cred krbmsg.KRBCred
			err = cred.Unmarshal(authCred.Deleg)
			if err != nil {
				k.client.logger.Info(message.Wrap(
					err,
					message.EAuthKerberosVerificationFailed,
					"Failed to marshal the KRB_CRED packet",
				))
				return mar2, k.principalUsername, false, nil
			}
			err = cred.DecryptEncPart(ticket.DecryptedEncPart.Key)
			if err != nil {
				return nil, "", false, err
			}

			cacheCreds, err := cred.ToCredentials()
			if err != nil {
				return nil, "", false, err
			}
			cache := credentials.CCacheFromCredentials(cacheCreds)
			mar, err := cache.Marshal()
			if err != nil {
				return nil, "", false, err
			}
			k.credentials = mar
		}

		return mar2, k.principalUsername, false, nil
	}

	return nil, "", false, fmt.Errorf("Invalid token")
}

// GSSAPIMicField is described in RFC4462 Section 3.5
type GSSAPIMicField struct {
	// SessionIdentifier is a random string identifying the ssh connection
	// (different from containerSSHs identifier)
	SessionIdentifier string
	// Request is the action that's being requested (50: SSH_MSG_USERAUTH_REQUEST)
	Request byte
	// UserName is the username that the user requests to log in as
	UserName string
	// Service is the service being used ('ssh-connection')
	Service string
	// Method is the authentication method in use ('gssapi-with-mic')
	Method string
}

func (mic *GSSAPIMicField) unmarshal(b []byte) error {
	err := ssh.Unmarshal(b, mic)

	if err != nil {
		return err
	}
	return nil
}

func (k *kerberosAuthContext) VerifyMIC(micField []byte, micToken []byte) error {
	var t gssapi.MICToken
	err := t.Unmarshal(micToken, false)
	if err != nil {
		return err
	}
	t.Payload = micField
	verified, err := t.Verify(k.key, keyusage.GSSAPI_INITIATOR_SIGN)
	if err != nil {
		return err
	}
	if !verified {
		return message.NewMessage(
			message.EAuthKerberosVerificationFailed,
			"Verify() returned unverified but no error",
		)
	}

	// MIC is verified, but need to ensure usernames match
	var field GSSAPIMicField
	err = field.unmarshal(micField)
	if err != nil {
		return err
	}

	if field.Request != internalSsh.SSH_MSG_USERAUTH_REQUEST || field.Service != "ssh-connection" || field.Method != "gssapi-with-mic" {
		return message.NewMessage(
			message.EAuthKerberosVerificationFailed,
			"Received MIC packet with unexpected values",
		)
	}

	if k.client.config.EnforceUsername && field.UserName != k.principalUsername {
		return message.UserMessage(
			message.EAuthKerberosVerificationFailed,
			"Unable to login with the requested username",
			"Cannot login to account %s using principal %s",
			field.UserName,
			k.principalUsername,
		)
	}

	k.loginUsername = field.UserName
	k.success = true

	return nil
}

func (k *kerberosAuthContext) DeleteSecContext() error {
	return nil
}

func (k *kerberosAuthContext) AllowLogin(
	username string,
	meta metadata.ConnectionAuthPendingMetadata,
) (metadata.ConnectionAuthenticatedMetadata, error) {
	if !k.Success() {
		return meta.AuthFailed(), nil
	}

	if k.loginUsername != username {
		return meta.AuthFailed(), nil
	}

	// Note: this is a redundant check as VerifyMIC already checks this
	// case, but it never hurts to be paranoid.
	if k.client.config.EnforceUsername && username != k.principalUsername {
		return meta.AuthFailed(), nil
	}

	return meta.Authenticated(username), nil
}
