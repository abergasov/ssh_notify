package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
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
	for {
		for line, prefix, err := r.ReadLine(); err != io.EOF; line, prefix, err = r.ReadLine() {
			if prefix {
				fmt.Fprint(out, string(line))
			} else {
				fmt.Fprintln(out, string(line))
			}
		}
		pos, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			panic(err)
		}
		for {
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
