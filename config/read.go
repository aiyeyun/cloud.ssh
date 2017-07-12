package config

import (
	"github.com/go-ini/ini"
	"os"
	"logger"
)

func Read(section, keys string) string {
	path := os.Getenv("GOPATH") + "/src/cloud.ssh/config/config.ini"
	cnf, err := ini.Load(path)
	if err != nil {
		logger.Log("ini File open failed")
		os.Exit(0)
	}

	section_name := section
	key_name := keys

	val := cnf.Section(section_name).Key(key_name).String()

	if val == "" {
		logger.Log(key_name + " not found")
	}
	return val
}