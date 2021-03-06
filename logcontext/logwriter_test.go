package logcontext_test

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/go-midway/midway/logcontext"
)

func TestLogWriter(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 1024))
	logger := logcontext.LogWriter(log.New(buf, "[testing] ", 0))
	fmt.Fprintf(logger, "hello logger")

	if want, have := "[testing] hello logger\n", strings.Trim(buf.String(), "\x00"); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}
