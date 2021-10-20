package test_test

import (
	"fmt"
	"testing"

	"github.com/containerssh/containerssh/internal/test"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
)

var krbConf = `
[libdefaults]
 dns_lookup_realm = false
 dns_lookup_kdc = false

[realms]
 %s = {
  kdc = 127.0.0.1:88
 }

[domain_realm]
`

func TestKerberos(t *testing.T) {
	krb := test.Kerberos(t)
	cfg := fmt.Sprintf(krbConf, krb.Realm())
	krbConfig, err := config.NewFromString(cfg)
	if err != nil {
		t.Fatalf("failed to parse Kerberos config (%v)", err)
	}
	cli := client.NewWithPassword(krb.AdminUsername(), krb.Realm(), krb.AdminPassword(), krbConfig)
	if err := cli.Login(); err != nil {
		t.Fatalf("failed to login (%v)", err)
	}
}
