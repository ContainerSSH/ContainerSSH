package ssh

import (
	"github.com/janoszen/containerssh/ssh/env"
	"github.com/janoszen/containerssh/ssh/pty"
	"github.com/janoszen/containerssh/ssh/request"
	"github.com/janoszen/containerssh/ssh/run"
	"github.com/janoszen/containerssh/ssh/signal"
	"github.com/janoszen/containerssh/ssh/window"
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
