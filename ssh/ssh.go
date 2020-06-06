package ssh

import (
	"containerssh/ssh/env"
	"containerssh/ssh/pty"
	"containerssh/ssh/request"
	"containerssh/ssh/shell"
	"containerssh/ssh/window"
)

func InitRequestHandlers() request.Handler {
	handler := request.NewHandler()
	handler.AddTypeHandler("env", env.SetEnvRequestTypeHandler)
	handler.AddTypeHandler("pty-req", pty.PtyRequestTypeHandler)
	handler.AddTypeHandler("shell", shell.ShellRequestTypeHandler)
	handler.AddTypeHandler("window-change", window.WindowRequestTypeHandler)
	return handler
}
