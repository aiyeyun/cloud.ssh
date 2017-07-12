package server

import (
	"golang.org/x/crypto/ssh"
	"net"
)

type CloudSSH struct {
	Addr    string
	Port    string
	User    string
	Pwd     string
	Columns uint32
	Rows    uint32
}

type ptyRequestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

func Telnet(addr, port, user, pwd string) error {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(pwd)},
	}

	client, err := ssh.Dial("tcp", addr+":"+port, config)
	defer func() {
		if err := recover(); err == "" {
			client.Close()
		}
	}()

	if err != nil {
		client.Close()
		return err
	}
	client.Close()
	return nil
}

func (cssh *CloudSSH) Connect() (*ssh.Client, ssh.Channel, error) {
	config := &ssh.ClientConfig{
		User: cssh.User,
		Auth: []ssh.AuthMethod{ssh.Password(cssh.Pwd)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", cssh.Addr + ":" + cssh.Port, config)
	if err != nil {
		return nil, nil, err
	}

	channel, _, err := client.Conn.OpenChannel("session", nil)
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO: 1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	var modeList []byte
	for k, v := range modes {
		kv := struct {
			Key byte
			Val uint32
		}{k, v}
		modeList = append(modeList, ssh.Marshal(&kv)...)
	}
	modeList = append(modeList, 0)

	req := ptyRequestMsg{
		Term:     "xterm",
		Columns:  cssh.Columns,
		Rows:     cssh.Rows,
		Width:    cssh.Columns * 8,
		Height:   cssh.Rows * 8,
		Modelist: string(modeList),
	}

	ok, err := channel.SendRequest("pty-req", true, ssh.Marshal(&req))
	if !ok || err != nil {
		channel.Close()
		client.Close()
		return nil, nil, err
	}

	ok, err = channel.SendRequest("shell", true, nil)
	if !ok || err != nil {
		channel.Close()
		client.Close()
		return nil, nil, err
	}

	return client, channel, err
}