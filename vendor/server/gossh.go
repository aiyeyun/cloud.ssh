package server

import (
	"golang.org/x/crypto/ssh"
	"log"
)

var ssh_client *ssh.Client

var ssh_chann ssh.Channel

type sshClient struct {
	user    string
	pwd     string
	addr    string
	client  *ssh.Client
}

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

func (sh *sshClient) Connect() (*sshClient, error) {
	config := &ssh.ClientConfig{
		User: sh.user,
		Auth: []ssh.AuthMethod{ssh.Password(sh.pwd)},
	}
	client, err := ssh.Dial("tcp", sh.addr, config)
	if err != nil {
		return nil, err
	}
	sh.client = client
	return sh, nil
}

func Start() (ssh.Channel) {
	client := new(sshClient)
	client.user = "root"
	client.pwd = "root"
	client.pwd = "whb1993822."
	client.addr = "123.56.225.185:22"
	sh_client, err := client.Connect()
	if err != nil {
		log.Println(err)
	}
	ssh_client = sh_client.client

	channel, _, err := ssh_client.Conn.OpenChannel("session", nil)
	if err != nil {
		panic(err)
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
		Columns:  204,
		Rows:     25,
		Width:    204 * 8,
		Height:   25 * 8,
		Modelist: string(modeList),
	}

	ok, err := channel.SendRequest("pty-req", true, ssh.Marshal(&req))
	if !ok || err != nil {
		log.Println(err)
		return nil
	}

	ok, err = channel.SendRequest("shell", true, nil)
	if !ok || err != nil {
		log.Println(err)
		return nil
	}

	return channel
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
		return err
	}
	return nil
}

func (cssh *CloudSSH) Connect() (*ssh.Client, ssh.Channel, error) {
	config := &ssh.ClientConfig{
		User: cssh.User,
		Auth: []ssh.AuthMethod{ssh.Password(cssh.Pwd)},
	}

	client, err := ssh.Dial("tcp", cssh.Addr + ":" + cssh.Port, config)
	if err != nil {
		client.Close()
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
		client.Close()
		return nil, nil, err
	}

	ok, err = channel.SendRequest("shell", true, nil)
	if !ok || err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, channel, err
}