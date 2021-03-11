module github.com/containerssh/containerssh

go 1.16

require (
	github.com/aws/aws-sdk-go v1.37.28
	github.com/containerssh/auditlog v0.9.9
	github.com/containerssh/auditlogintegration v0.9.4
	github.com/containerssh/auth v0.9.6
	github.com/containerssh/authintegration v0.9.4
	github.com/containerssh/backend v0.9.8
	github.com/containerssh/configuration v0.9.9
	github.com/containerssh/geoip v0.9.4
	github.com/containerssh/http v0.9.9
	github.com/containerssh/log v0.9.13
	github.com/containerssh/metrics v0.9.8
	github.com/containerssh/metricsintegration v0.9.3
	github.com/containerssh/service v0.9.3
	github.com/containerssh/sshserver v0.9.19
	github.com/containerssh/structutils v0.9.0
	github.com/cucumber/godog v0.11.0
	github.com/docker/docker v20.10.5+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.3 // indirect
	github.com/oschwald/maxminddb-golang v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/oauth2 v0.0.0-20210113205817-d3ed898aa8a3 // indirect
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210114201628-6edceaf6022f // indirect
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect
)

replace (
	// Fixes CVE-2020-9283
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2 => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975 => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	// Fixes CVE-2020-14040
	golang.org/x/text v0.3.0 => golang.org/x/text v0.3.3
	golang.org/x/text v0.3.1 => golang.org/x/text v0.3.3
	golang.org/x/text v0.3.2 => golang.org/x/text v0.3.3
)

// Exclude this package because it got renamed to /moby/ which breaks packages.
exclude github.com/docker/spdystream v0.2.0
