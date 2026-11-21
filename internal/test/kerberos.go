package test

import (
	"embed"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

var krbLock = &sync.Mutex{}

//go:embed krb
var krbBuildRoot embed.FS

// Kerberos starts a new test Kerberos instance. This Kerberos instance is available on the host system. The
// realm can be obtained from the returned helper object and will be accessible via public DNS. However, to decrease
// the need to rely on DNS working for your tests you should set up the KDC address explicitly, for example:
//
//     var krbConfigTemplate = `[libdefaults]
//     dns_lookup_realm = false
//     dns_lookup_kdc = false
//
//     [realms]
//     %s = {
//         kdc = %s:%d
//     }
//
//     [domain_realm]`
//
//     func TestMe(t *testing.T) {
//         krb := test.Kerberos(t)
//         krbConfig := fmt.Sprintf(
//             krbConfigTemplate,
//             krb.Realm(),
//             krb.KDCHost(),
//             krb.KDCPort(),
//         )
//
//         // Do more Kerberos stuff
//     }
func Kerberos(t *testing.T) KerberosHelper {
	t.Helper()
	krbLock.Lock()
	t.Cleanup(
		func() {
			krbLock.Unlock()
		})
	adminUsername := "admin"
	adminPassword := "testing"

	files := dockerBuildRootFiles(krbBuildRoot, "krb")
	cnt := containerFromBuild(
		t,
		"krb",
		files,
		nil,
		[]string{
			fmt.Sprintf("KERBEROS_USERNAME=%s", adminUsername),
			fmt.Sprintf("KERBEROS_PASSWORD=%s", adminPassword),
		},
		map[string]string{
			"88/tcp":  "88",
			"88/udp":  "88",
			"464/tcp": "464",
			"464/udp": "464",
			"750/tcp": "750",
			"750/udp": "750",
		},
	)

	helper := &kerberosHelper{
		t:             t,
		cnt:           cnt,
		adminUsername: adminUsername,
		adminPassword: adminPassword,
	}

	helper.wait()

	return helper
}

// KerberosHelper is a helper which contains the information about the running Kerberos test service. The individual
// functions can be used to obtain kerberos parameters.
type KerberosHelper interface {
	// Realm returns the Kerberos realm to authenticate with.
	Realm() string
	// KDCHost returns the IP address of the KDC service.
	KDCHost() string
	// KDCPort returns the port number the KDC service is running on.
	KDCPort() int
	// AdminUsername returns a Kerberos username which has admin privileges.
	AdminUsername() string
	// AdminPassword returns the password for the admin username from AdminUsername.
	AdminPassword() string
	// Get service keytab
	Keytab() []byte
}

type kerberosHelper struct {
	adminUsername string
	adminPassword string
	cnt           container
	t             *testing.T
}

func (k *kerberosHelper) KDCHost() string {
	return "127.0.0.1"
}

func (k *kerberosHelper) Realm() string {
	return "TESTING.CONTAINERSSH.IO"
}

func (k *kerberosHelper) AdminUsername() string {
	return k.adminUsername
}

func (k *kerberosHelper) AdminPassword() string {
	return k.adminPassword
}

func (k *kerberosHelper) KDCPort() int {
	return 88
}

func (k *kerberosHelper) Keytab() []byte {
	return k.cnt.extractFile("/test.keytab")
}

func (k *kerberosHelper) wait() {
	k.t.Log("Waiting for the KDC to come up...")
	tries := 0
	sleepTime := 5
	for {
		if tries > 30 {
			k.t.Fatalf("The KDC failed to come up in %d seconds", sleepTime*30)
		}
		sock, err := net.Dial("tcp", "127.0.0.1:88")
		time.Sleep(time.Duration(sleepTime) * time.Second)
		if err != nil {
			tries++
		} else {
			_ = sock.Close()

			return
		}
	}
}
