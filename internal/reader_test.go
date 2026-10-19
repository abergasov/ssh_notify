package internal

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTailUntilRotateProcessesReplacementAfterRotationGap(t *testing.T) {
	directory := t.TempDir()
	fileName := filepath.Join(directory, "auth.log")
	historical := "sshd[1]: Accepted publickey for old from 10.0.0.1 port 22\n"
	if err := ioutil.WriteFile(fileName, []byte(historical), 0600); err != nil {
		t.Fatal(err)
	}

	oldConf, oldOutput, oldFlags := conf, log.Writer(), log.Flags()
	defer func() {
		conf = oldConf
		log.SetOutput(oldOutput)
		log.SetFlags(oldFlags)
	}()
	conf = &Config{KnownIps: make(map[string]string)}
	var output bytes.Buffer
	log.SetOutput(&output)
	log.SetFlags(0)

	skipRows := true
	done := make(chan struct{})
	go func() {
		tailUntilRotate(fileName, &skipRows)
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	if err := os.Rename(fileName, filepath.Join(directory, "auth.log.1")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("tailUntilRotate() did not return while the replacement file was absent")
	}
	if skipRows {
		t.Fatal("tailUntilRotate() kept skipping records after the original file reached EOF")
	}

	replacement := "sshd[2]: Accepted publickey for new from 10.0.0.2 port 22\n"
	if err := ioutil.WriteFile(fileName, []byte(replacement), 0600); err != nil {
		t.Fatal(err)
	}
	done = make(chan struct{})
	go func() {
		tailUntilRotate(fileName, &skipRows)
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	if err := os.Rename(fileName, filepath.Join(directory, "auth.log.2")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("tailUntilRotate() did not return after the replacement rotated")
	}

	got := output.String()
	if !strings.Contains(got, "new from 10.0.0.2") {
		t.Fatalf("replacement record was not processed: %s", got)
	}
	if strings.Contains(got, "old from 10.0.0.1") {
		t.Fatalf("startup history was processed: %s", got)
	}
}
