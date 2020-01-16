package app

type Config struct {
	SSHLogFile         string
	TelegramBotToken   string
	TelegramNotifyChat string
}

var conf *Config

func New() *Config {
	conf = &Config{
		SSHLogFile:         "/var/log/auth.log",
		TelegramBotToken:   "213123123",
		TelegramNotifyChat: "123123123123",
	}
	return conf
}
