module github.com/containerssh/containerssh

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.33
	github.com/containerssh/auditlog v1.0.0
	github.com/containerssh/auditlogintegration v1.0.0
	github.com/containerssh/auth v1.0.0
	github.com/containerssh/authintegration v1.0.0
	github.com/containerssh/backend v1.0.0
	github.com/containerssh/configuration v1.0.0
	github.com/containerssh/geoip v1.0.0
	github.com/containerssh/http v1.0.0
	github.com/containerssh/log v1.0.0
	github.com/containerssh/metrics v1.0.0
	github.com/containerssh/metricsintegration v1.0.0
	github.com/containerssh/service v1.0.0
	github.com/containerssh/sshserver v1.0.0
	github.com/containerssh/structutils v1.0.0
	github.com/cucumber/godog v0.11.0
	github.com/docker/docker v20.10.5+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/go-enry/go-license-detector/v4 v4.1.0
	github.com/mitchellh/golicense v0.2.0
	github.com/rsc/goversion v1.2.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
)

// Exclude this package because it got renamed to /moby/ which breaks packages.
exclude github.com/docker/spdystream v0.2.0

// Fixes CVE-2020-9283
replace (
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2 => golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 => golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975 => golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 => golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
)

// Fixes CVE-2020-14040
replace (
	golang.org/x/text v0.3.0 => golang.org/x/text v0.3.3
	golang.org/x/text v0.3.1 => golang.org/x/text v0.3.3
	golang.org/x/text v0.3.2 => golang.org/x/text v0.3.3
)

// Fixes CVE-2019-11254
replace (
	gopkg.in/yaml.v2 v2.2.0 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.1 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.2 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.3 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.4 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.5 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.6 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.7 => gopkg.in/yaml.v2 v2.2.8
)
