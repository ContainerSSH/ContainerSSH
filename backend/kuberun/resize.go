package kuberun

import "k8s.io/client-go/tools/remotecommand"

func (session *kubeRunSession) Resize(cols uint, rows uint) error {
	//Todo this could be done nicer
	go func() {
		session.terminalSizeQueue.resizeChan <- remotecommand.TerminalSize{
			Width:  uint16(cols),
			Height: uint16(rows),
		}
	}()
	return nil
}
