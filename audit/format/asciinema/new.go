package asciinema

import (
	"fmt"
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/log"
)

func NewEncoder(_ log.Logger) (audit.Encoder, error) {
	return nil, fmt.Errorf("not implemented")
}
