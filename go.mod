module github.com/containerssh/containerssh

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.47
	github.com/containerssh/auditlog v1.0.0
	github.com/containerssh/auditlogintegration/v2 v2.0.0
	github.com/containerssh/auth v1.0.1
	github.com/containerssh/authintegration/v2 v2.0.0
	github.com/containerssh/backend/v3 v3.0.0
	github.com/containerssh/configuration/v3 v3.0.0
	github.com/containerssh/geoip v1.0.0
	github.com/containerssh/health v1.1.0
	github.com/containerssh/http v1.2.0
	github.com/containerssh/log v1.1.6
	github.com/containerssh/metrics v1.0.0
	github.com/containerssh/metricsintegration v1.0.0
	github.com/containerssh/service v1.0.0
	github.com/containerssh/sshserver/v2 v2.0.0
	github.com/containerssh/structutils v1.1.0
	github.com/cucumber/godog v0.11.0
	github.com/docker/docker v20.10.6+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/go-enry/go-license-detector/v4 v4.1.0
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/mitchellh/golicense v0.2.0
	github.com/rsc/goversion v1.2.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/sys v0.0.0-20210525143221-35b2ab0089ea // indirect
	google.golang.org/genproto v0.0.0-20210524171403-669157292da3 // indirect
)

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

replace github.com/containerssh/backend/v3 v3.0.0 => ../backend

replace github.com/containerssh/metricsintegration v1.0.0 => ../metricsintegration

replace github.com/containerssh/configuration/v3 v3.0.0 => ../configuration

replace github.com/containerssh/docker/v3 v3.0.0 => ../docker

replace github.com/containerssh/kubernetes/v3 v3.0.0 => ../kubernetes

replace github.com/containerssh/sshproxy/v2 v2.0.0 => ../sshproxy

replace github.com/containerssh/sshserver/v2 v2.0.0 => ../sshserver

replace github.com/containerssh/auditlogintegration/v2 v2.0.0 => ../auditlogintegration

replace github.com/containerssh/authintegration/v2 v2.0.0 => ../authintegration

replace github.com/containerssh/auth/v2 v2.0.0 => ../auth

replace github.com/containerssh/security/v2 v2.0.0 => ../security
