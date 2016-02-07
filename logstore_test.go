package logstore

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestAll(t *testing.T) {
	l := New("_test/", 10)
	for i := 0; i < 20; i++ {
		out, _ := exec.Command("uuidgen").Output()
		uuid := strings.Trim(string(out), "\n")
		n, err := l.Append(uuid, []byte(uuid))
		if err != nil {
			t.Error(err)
		}
		fmt.Println("Write", n)
	}
	b := make([]byte, 2024)
	n, err := l.Read(0, int64(0), b)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Read bytes", n)
	fmt.Println("Buffer", b)
	fmt.Println("Buffer Size", len(b))
	msgs, err := l.UnMarshal(b[:n])
	if len(msgs) > 0 {
		fmt.Println(string(msgs[0].Payload))
	}
}
