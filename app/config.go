package app

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
)

type Config struct {
	SSHLogFile         string
	ServerName         string
	TelegramBotToken   string
	TelegramNotifyChat string
}

var conf *Config

func New() *Config {
	parseredConf := readConf()
	conf = &Config{
		SSHLogFile:         getVariableOrDefault(parseredConf, "SSHLogFile", "/var/log/auth.log"),
		TelegramBotToken:   getVariableOrDefault(parseredConf, "TelegramBotToken", ""),
		TelegramNotifyChat: getVariableOrDefault(parseredConf, "TelegramNotifyChat", ""),
		ServerName:         getVariableOrDefault(parseredConf, "ServerName", "default_server_name"),
	}
	log.Print("Config loaded")
	log.Print("Log file", conf.SSHLogFile)
	log.Print("Server name", conf.ServerName)
	log.Print("Telegram chat", conf.TelegramNotifyChat)
	return conf
}

func readConf() *map[string]string {
	file, err := os.Open("/etc/ssh_notify.conf")
	if err != nil {
		log.Print("Can't open app config file /etc/ssh_notify.conf")
		panic(err.Error())
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	tmpConf := map[string]string{}
	for {
		line, err := reader.ReadString('\n')

		equal := strings.Index(line, "=")
		if equal == -1 {
			continue
		}
		if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
			if len(line) <= equal {
				continue
			}
			tmpConf[key] = strings.TrimSpace(line[equal+1:])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err.Error())
		}
	}
	return &tmpConf
}

func getVariableOrDefault(tmpConf *map[string]string, name string, defaultValue string) string {
	for key, val := range *tmpConf {
		if key == name {
			return val
		}
	}
	return defaultValue
}
