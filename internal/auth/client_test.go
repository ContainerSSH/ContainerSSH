package auth_test

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/containerssh/auth"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/message"
	"github.com/containerssh/geoip"
	"github.com/containerssh/http"
	"github.com/containerssh/metrics"
)

// TestPasswordDisabled tests if the call fails with the correct error if the password authentication method is
// disabled. The inverse is not tested because it is already tested by the integration test.
func TestPasswordDisabled(t *testing.T) {
	config := auth.ClientConfig{
		ClientConfiguration: http.ClientConfiguration{
			URL:            "http://localhost:8080",
			AllowRedirects: false,
			Timeout:        100 * time.Millisecond,
		},
		AuthTimeout: time.Second,
		Password:    false,
		PubKey:      true,
	}
	geoIp, err := geoip.New(geoip.Config{
		Provider: geoip.DummyProvider,
	})
	if err != nil {
		t.Fatal(err)
	}
	c, err := auth.NewHttpAuthClient(
		config,
		log.NewTestLogger(t),
		metrics.New(geoIp),
	)
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.Password("foo", []byte("bar"), "asdf", net.ParseIP("127.0.0.1"))
	if result {
		t.Fatal("Password authentication method resulted in successful authentication.")
	}
	if err == nil {
		t.Fatal("Password authentication method did not result in an error.")
	}
	var realErr message.Message
	if !errors.As(err, &realErr) {
		t.Fatal("Password authentication did not return a log.Message")
	}
	if realErr.Code() != auth.EDisabled {
		t.Fatal("Disabled password authentication did not return an auth.EDisabled code.")
	}
}

// TestPubKeyDisabled tests if the call fails with the correct error if the public key authentication method is
// disabled. The inverse is not tested because it is already tested by the integration test.
func TestPubKeyDisabled(t *testing.T) {
	config := auth.ClientConfig{
		ClientConfiguration: http.ClientConfiguration{
			URL:            "http://localhost:8080",
			AllowRedirects: false,
			Timeout:        100 * time.Millisecond,
		},
		AuthTimeout: time.Second,
		Password:    true,
		PubKey:      false,
	}
	geoIp, err := geoip.New(geoip.Config{
		Provider: geoip.DummyProvider,
	})
	if err != nil {
		t.Fatal(err)
	}
	c, err := auth.NewHttpAuthClient(
		config,
		log.NewTestLogger(t),
		metrics.New(geoIp),
	)
	if err != nil {
		t.Fatal(err)
	}
	result, err := c.PubKey(
		"foo",
		"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDP39LqSomHi4kicGADA3XVQoYxzNkvrBLOqN5AEEP01p0TZ39LXa6FdB4Pmvg8h51c+BNLoxpYrTk4UibMD87OPKYYXrNmLvq0GwjMPYpzoICevAJm+/2sDVlK9sXT93Fkin+tei+Evgf/hQK0xN+HXqP8dz8SWSXeWjBv588eHHCdrV+0FlZLXH+9D18tD4BNPHe9iJLpeeH6gsvQBvArXcIEQVvHIo1cCcsy28ymUFndG55LdOaTCA+pcfHLmRtL8HO2mI2Qc/0HBSc2d1gb3lHAnmdMT82K58OjRp9Tegc5hVuKVE+hkmNjfo3f1mVHsNu6JYLxRngnbJ20QdzuKcPb3pRMty+ggRgEQExvgl1pC3OVcgyc8YX1eXiyhYy0kXT/Jg++AcaIC1Xk/hDfB0T7WxCO0Wwd4KSjKr79tIxM/m4jP2K1Hk4yAnT7mZQ0GjdphLLuDk3yt8R809SPuzkPCXBM0sL6FrqT2GVDNihN2pBh1MyuUt7S8ZXpuW0=",
		"asdf",
		net.ParseIP("127.0.0.1"),
	)
	if result {
		t.Fatal("Public key authentication method resulted in successful authentication.")
	}
	if err == nil {
		t.Fatal("Public key authentication method did not result in an error.")
	}
	var realErr message.Message
	if !errors.As(err, &realErr) {
		t.Fatal("Public key authentication did not return a log.Message")
	}
	if realErr.Code() != auth.EDisabled {
		t.Fatal("Disabled public key authentication did not return an auth.EDisabled code.")
	}
}
