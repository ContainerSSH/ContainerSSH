package config_test

import (
	"reflect"
	"strings"
	"testing"

    configuration "go.containerssh.io/libcontainerssh/config"
)

// TestOperatorCompatibility is a test that tries to create compatibility with the Kubernetes operator
// SDK. We cannot create full compatibility because several external libraries don't adhere to this standard.
// Therefore, we exclude the Kubernetes and Docker backends from this check.
func TestOperatorCompatibility(t *testing.T) {
	appConfig := configuration.AppConfig{}
	v := reflect.TypeOf(appConfig)
	verify(v, t, []string{"AppConfig"})
}

func verify(v reflect.Type, t *testing.T, path []string) {
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			tag := v.Field(i).Tag
			jsonValue, ok := tag.Lookup("json")
			if !ok {
				t.Errorf(
					"Config field has no JSON tag set: %s",
					strings.Join(append(path, v.Field(i).Name), " -> "))
			}
			if jsonValue != "-" &&
				v.Field(i).Name != "Docker" &&
				v.Field(i).Name != "DockerRun" &&
				v.Field(i).Name != "Kubernetes" &&
				v.Field(i).Name != "KubeRun" {
				// We exclude Docker since it
				verify(v.Field(i).Type, t, append(path, v.Field(i).Name))
			}
		}
	case reflect.Slice:
		verify(v.Elem(), t, path)
	case reflect.Ptr:
		verify(v.Elem(), t, path)
	}
}
