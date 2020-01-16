package main

import (
	"os"
	"ssh_notify/app"
)

var conf = app.New()

func main() {
	if !fileExist(conf.SSHLogFile) {
		println("ssh log file not found")
		println("check file exists at", conf.SSHLogFile)
		println("also you can set another file through the config")
		return
	}
	if !fileHasReadPermissions(conf.SSHLogFile) {
		println("app has't read permissions to file", conf.SSHLogFile)
		println("make sure that the application starts as root")
		return
	}
	app.Tail(conf.SSHLogFile, os.Stdout)
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
		println("error while open log file")
		println(err.Error())
		return false
	}
	return true
}
