package internal

import (
	"bufio"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	ipReg     = regexp.MustCompile(`([0-9]{1,3}[\.]){3}[0-9]{1,3}`)
	acceptReg = regexp.MustCompile(`sshd[[0-9]+]:.Accepted`)
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
		LogMessage(err.Error(), "impossible open log file", fileName)
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	info, err := f.Stat()
	if err != nil {
		LogMessage(err.Error(), "impossible get init file info", fileName)
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
				LogMessage(err.Error(), "impossible get file info", fileName)
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
	//inoOld := getFileIno(info)
	//inoNew := getFileIno(infoNew)
	return !os.SameFile(info, infoNew)
}

func searchMatch(row string) {
	if matched := acceptReg.MatchString(row); !matched {
		return
	}
	log.Print("Found matches in auth log", row)
	var titleMsg string
	if len(conf.KnownIps) > 0 {
		titleMsg = "UNKNOWN IP login on server"
		if ipReg.MatchString(row) {
			connectedIp := strings.TrimSpace((ipReg.FindAllString(row, -1))[0])
			if knownServerName, ok := conf.KnownIps[connectedIp]; ok {
				titleMsg = knownServerName + " login on server"
			}
		}
	} else {
		titleMsg = "New server login"
	}
	LogMessage(titleMsg, conf.ServerName, row)
}
