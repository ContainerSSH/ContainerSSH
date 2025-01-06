<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Test Helper Library</h1>

This library helps with bringing up services for testing, such as S3, oAuth, etc. **All services require an exposed Docker socket to work.**

## Starting an S3 service

To start the S3 service and then use it with the AWS client as follows:

```go
package your_test

import (
    "testing"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/containerssh/test"
)

func TestYourFunc(t *testing.T) {
    s3Service := test.S3(t)

    awsConfig := &aws.Config{
        Credentials: credentials.NewCredentials(
            &credentials.StaticProvider{
                Value: credentials.Value{
                    AccessKeyID:     s3Service.AccessKey(),
                    SecretAccessKey: s3Service.SecretKey(),
                },
            },
        ),
        Endpoint:         aws.String(s3Service.URL()),
        Region:           aws.String(s3Service.Region()),
        S3ForcePathStyle: aws.Bool(s3Service.PathStyle()),
    }
    sess, err := session.NewSession(awsConfig)
    if err != nil {
        t.Fatalf("failed to establish S3 session (%v)", err)
    }
    s3Connection := s3.New(sess)
    
    // ...
}
```

That's it! Now you have a working S3 connection for testing!

## Starting the Kerberos service

You can start the Kerberos service and then use it to authenticate like this:

```go
package your_test

import (
    "fmt"
    "testing"

    "github.com/containerssh/test"
    "github.com/containerssh/gokrb5/v8/client"
    "github.com/containerssh/gokrb5/v8/config"
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
    
    krbConfig, err := config.NewFromString(fmt.Sprintf(krbConf, krb.Realm()))
    if err != nil {
        t.Fatalf("invalid Kerberos config (%v)", err)
    }
    cli := client.NewWithPassword(
        krb.AdminUsername(),
        krb.Realm(),
        krb.AdminPassword(),
        krbConfig,
    )
    if err := cli.Login(); err != nil {
        t.Fatalf("failed to login (%v)", err)
    }
}
```

**⚠️ Warning!** The Kerberos server image is built locally and may take several minutes to build!

**👉 Tip:** The Kerberos service works locally without DNS lookups, but can also use the test DNS records published under `TESTING.CONTAINERSSH.IO`. It is recommended that test cases work without DNS lookups.

**👉 Tip:** You can use the contents of the [krb](krb) directory to build and start a Kerberos service for testing purposes like this:

```bash
docker build -t krb .
docker run \
       --rm \
      -p 127.0.0.1:88:88 \
      -p 127.0.0.1:88:88/udp \
      -e KERBEROS_USERNAME=admin \
      -e KERBEROS_PASSWORD=testing \
      -ti \
      krb
```
