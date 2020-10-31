package asciinema

import (
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/log"
)

func NewEncoder(logger log.Logger) (audit.Encoder, error) {
	return &encoder{
		logger: logger,
	}, nil
}
