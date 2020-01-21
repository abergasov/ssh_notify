package main

import (
	"log"
	"os"
	"ssh_notify/app"
)

var conf = app.New()

func main() {
	if !fileExist(conf.SSHLogFile) {
		log.Print("ssh log file not found")
		log.Print("check file exists at", conf.SSHLogFile)
		log.Print("also you can set another file through the config")
		return
	}
	if !fileHasReadPermissions(conf.SSHLogFile) {
		log.Print("app has't read permissions to file", conf.SSHLogFile)
		log.Print("make sure that the application starts as root")
		return
	}
	log.Print("Log file ok, start watch", conf.SSHLogFile)
	app.Tail(conf.SSHLogFile)
}

func fileExist(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func fileHasReadPermissions(fileName string) bool {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		if os.IsPermission(err) {
			return false
		}
		log.Print("error while open log file")
		log.Print(err.Error())
		return false
	}
	return true
}
