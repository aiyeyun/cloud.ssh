package server

import (
	"logo"
	"session"
)

func Run()  {
	go logo.Show()
	Socket()
}

var globalSSHSessions *session.SSHSessionManage
func init()  {
	globalSSHSessions, _ = session.NewSSHSessionManage("cloud.ssh.sid")
}