package ssh

import (
	"containerssh/ssh/env"
	"containerssh/ssh/pty"
	"containerssh/ssh/request"
	"containerssh/ssh/run"
	"containerssh/ssh/signal"
	"containerssh/ssh/window"
)

func InitRequestHandlers() request.Handler {
	handler := request.NewHandler()
	handler.AddTypeHandler("env", env.RequestTypeHandler)
	handler.AddTypeHandler("pty-req", pty.RequestTypeHandler)
	handler.AddTypeHandler("shell", run.ShellRequestTypeHandler)
	handler.AddTypeHandler("exec", run.ExecRequestTypeHandler)
	handler.AddTypeHandler("window-change", window.RequestTypeHandler)
	handler.AddTypeHandler("signal", signal.RequestTypeHandler)
	return handler
}
