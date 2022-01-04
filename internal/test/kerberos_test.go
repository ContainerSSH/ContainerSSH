package test_test

import (
	"fmt"
	"testing"

	"github.com/containerssh/libcontainerssh/internal/test"
	"github.com/containerssh/gokrb5/v8/client"
	"github.com/containerssh/gokrb5/v8/config"
)

var krbConf = `
[libdefaults]
 dns_lookup_realm = false
 dns_lookup_kdc = false

[realms]
 %s = {
  kdc = %s:%d
 }

[domain_realm]
`

func TestKerberos(t *testing.T) {
	krb := test.Kerberos(t)
	cfg := fmt.Sprintf(krbConf, krb.Realm(), krb.KDCHost(), krb.KDCPort())
	krbConfig, err := config.NewFromString(cfg)
	if err != nil {
		t.Fatalf("failed to parse Kerberos config (%v)", err)
	}
	cli := client.NewWithPassword(krb.AdminUsername(), krb.Realm(), krb.AdminPassword(), krbConfig)
	if err := cli.Login(); err != nil {
		t.Fatalf("failed to login (%v)", err)
	}
}
