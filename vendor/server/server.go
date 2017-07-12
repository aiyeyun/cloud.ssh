package server

import (
	"logo"
	"session"
)

func Run()  {
	go logo.Show()
	Socket()
}

var globalSessions *session.Manager
func init()  {
	globalSessions, _ = session.NewSessionManager("memory", "cloud.ssh.sid", 3600)
	go globalSessions.GC()
}