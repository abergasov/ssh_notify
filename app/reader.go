package app

import (
	"bufio"
	"crypto/sha256"
	"io"
	"log"
	"os"
	"regexp"
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
				if checkFileMoved(fileName, hashFile(f)) {
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

func checkFileMoved(fileName string, currentHash string) bool {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	return hashFile(f) != currentHash
}

func hashFile(f *os.File) string {
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}
	return string(h.Sum(nil))
}

func searchMatch(row string) {
	matched, _ := regexp.MatchString(`sshd\[[0-9]{1,}\]:.Accepted`, row)
	if !matched {
		return
	}
	log.Print("Found matches in auth log", row)
	go LogMessage("New server login", []string{conf.ServerName, row})
}
