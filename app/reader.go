package app

import (
	"bufio"
	"io"
	"log"
	"os"
	"regexp"
	"time"
)

func Tail(filename string, out io.Writer) {
	f, err := os.Open(filename)
	if err != nil {
		LogMessage(err.Error(), []string{"impossible open log file", filename})
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	info, err := f.Stat()
	if err != nil {
		LogMessage(err.Error(), []string{"impossible get init file info", filename})
		return
	}
	oldSize := info.Size()
	skipRows := true
	for {
		for line, _, err := r.ReadLine(); err != io.EOF; line, _, err = r.ReadLine() {
			if skipRows {
				continue
			}
			searchMatch(string(line))
		}
		pos, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			panic(err)
		}
		for {
			skipRows = false
			time.Sleep(time.Second)
			newInfo, err := f.Stat()
			if err != nil {
				LogMessage(err.Error(), []string{"impossible get file info", filename})
				return
			}
			newSize := newInfo.Size()
			if newSize == oldSize {
				continue
			}
			if newSize < oldSize {
				f.Seek(0, 0)
			} else {
				f.Seek(pos, io.SeekStart)
			}
			r = bufio.NewReader(f)
			oldSize = newSize
			break
		}
	}
}

func searchMatch(row string) {
	matched, _ := regexp.MatchString(`sshd\[[0-9]{1,}\]:.Accepted`, row)
	if !matched {
		return
	}
	log.Print("Found matches in auth log", row)
	go LogMessage("New server login", []string{conf.ServerName, row})
}
