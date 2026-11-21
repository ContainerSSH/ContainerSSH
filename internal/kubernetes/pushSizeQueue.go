package kubernetes

import (
	"context"

	"k8s.io/client-go/tools/remotecommand"
)

type pushSizeQueue interface {
	remotecommand.TerminalSizeQueue

	Push(context.Context, remotecommand.TerminalSize) error
	Stop()
}
