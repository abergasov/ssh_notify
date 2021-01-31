package internal

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
	SlackToken         string
	SlackChannel       string
	KnownIps           map[string]string
}

var conf *Config

func New() *Config {
	parseredConf := readConf()
	conf = &Config{
		SSHLogFile:         getVariableOrDefault(parseredConf, "SSHLogFile", "/var/log/auth.log"),
		TelegramBotToken:   getVariableOrDefault(parseredConf, "TelegramBotToken", ""),
		TelegramNotifyChat: getVariableOrDefault(parseredConf, "TelegramNotifyChat", ""),
		ServerName:         getVariableOrDefault(parseredConf, "ServerName", "default_server_name"),
		SlackToken:         getVariableOrDefault(parseredConf, "SlackBotToken", ""),
		SlackChannel:       getVariableOrDefault(parseredConf, "SlackTargetChannel", ""),
		KnownIps:           make(map[string]string),
	}

	ips := getVariableOrDefault(parseredConf, "KnownIps", "")
	if len(ips) != 0 {
		servers := strings.Split(ips, ";")
		for _, str := range servers {
			srv := strings.Split(strings.TrimSpace(str), ":")
			if len(srv) != 2 {
				continue
			}
			conf.KnownIps[strings.TrimSpace(srv[0])] = strings.TrimSpace(srv[1])
		}
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
		log.Print("Can't open internal config file /etc/ssh_notify.conf")
		panic(err.Error())
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	tmpConf := map[string]string{}
	for {
		line, err := reader.ReadString('\n')

		if err == io.EOF {
			break
		}

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
