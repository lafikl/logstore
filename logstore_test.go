package logstore

import (
	"os/exec"
	"strings"
	"testing"
)

func TestAll(t *testing.T) {
	l := New("_test/", 10)
	for i := 0; i < 20; i++ {
		out, _ := exec.Command("uuidgen").Output()
		uuid := strings.Trim(string(out), "\n")
		l.Append(uuid, []byte("hello world!"))
	}
}
