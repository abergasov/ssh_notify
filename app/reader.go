package app

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"syscall"
	"time"
)

func Tail(filename string) {
	skipRows := true
	for {
		tailUntilRotate(filename, &skipRows)
	}
}

func tailUntilRotate(fileName string, skipRows *bool) {
	f, err := os.Open(fileName)
	if err != nil {
		LogMessage(err.Error(), []string{"impossible open log file", fileName})
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	info, err := f.Stat()
	if err != nil {
		LogMessage(err.Error(), []string{"impossible get init file info", fileName})
		return
	}
	oldSize := info.Size()
	for {
		for line, _, err := r.ReadLine(); err != io.EOF; line, _, err = r.ReadLine() {
			if *skipRows {
				continue
			}
			searchMatch(string(line))
		}
		pos, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			panic(err)
		}
		for {
			*skipRows = false
			time.Sleep(time.Second)
			newInfo, err := f.Stat()
			if err != nil {
				LogMessage(err.Error(), []string{"impossible get file info", fileName})
				return
			}
			newSize := newInfo.Size()
			if newSize == oldSize {
				if checkFileMoved(fileName, info) {
					println("files are not equal, reopen file")
					return
				}
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

func checkFileMoved(fileName string, info os.FileInfo) bool {
	//println("checking file changed")
	ff, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer ff.Close()
	infoNew, _ := ff.Stat()
	return getFileIno(info) != getFileIno(infoNew)
}

func getFileIno(info os.FileInfo) uint64 {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		fmt.Printf("Not a syscall.Stat_t")
	}
	return stat.Ino
}

func searchMatch(row string) {
	matched, _ := regexp.MatchString(`sshd\[[0-9]{1,}\]:.Accepted`, row)
	if !matched {
		return
	}
	log.Print("Found matches in auth log", row)
	go LogMessage("New server login", []string{conf.ServerName, row})
}
