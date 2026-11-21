package config

import (
	"fmt"
	"strings"
)

func validateEnum(value string, enum []string) error {
	for _, v := range enum {
		if v == value {
			return nil
		}
	}
	return fmt.Errorf("invalid value: %s, expected one of: %s", value, strings.Join(enum, ", "))
}
