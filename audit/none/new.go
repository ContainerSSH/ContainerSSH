package none

import "github.com/containerssh/containerssh/audit"

func New() audit.Plugin {
	return &Plugin{}
}
