package kuberun

import (
	"fmt"
)

func (session *kubeRunSession) SendSignal(_ string) error {
	return fmt.Errorf("sending signals to kubernetes pods is currently not supported (issue #24957)")
}
