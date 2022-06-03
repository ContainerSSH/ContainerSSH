package geoip_test

import (
	"net"
	"os"
	"path"
	"testing"

    "go.containerssh.io/libcontainerssh/config"
	"github.com/stretchr/testify/assert"

    "go.containerssh.io/libcontainerssh/internal/geoip"
)

func TestMaxMind(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Skipf("failed to get current working directory, skipping test (%v)", err)
		return
	}
	geoIpFile := path.Join(dir, "data", "test-data", "GeoIP2-Country-Test.mmdb")
	if _, err := os.Stat(geoIpFile); err != nil {
		t.Skipf("mmdb test file doesn't exist, skipping test (%v)", err)
		return
	}
	provider, err := geoip.New(
		config.GeoIPConfig{
			Provider:   "maxmind",
			GeoIP2File: geoIpFile,
		},
	)
	assert.Nil(t, err, "failed to create GeoIP lookup")
	assert.Equal(t, "RU", provider.Lookup(net.ParseIP("2a02:efc0::1")))
	assert.Equal(t, "XX", provider.Lookup(net.ParseIP("127.0.0.1")))
}
