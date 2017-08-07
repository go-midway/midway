package logcontext_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-midway/midway/logcontext"
)

func TestApplyLogger(t *testing.T) {
	srv := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logcontext.GetLogger(r.Context())
		if logger != nil {
			logger.Log("hello", "world")
			fmt.Fprintf(w, "success")
			return
		}
		t.Logf("got no log context")
		fmt.Fprintf(w, "failed")
	}))
	buf := bytes.NewBuffer(make([]byte, 256))
	newLogger := func() kitlog.Logger {
		return kitlog.NewLogfmtLogger(buf)
	}
	srv = logcontext.ApplyLogger(newLogger)(srv)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://foobar.com/hello/world", nil)
	r.RemoteAddr = "http://somewhere.com:1234"
	r.Header.Add("X-Request-ID", "helloid")
	srv.ServeHTTP(w, r)

	if want, have := `request_id=helloid at=info method=GET path=/hello/world protocol=http remote_addr=http://somewhere.com:1234`+"\nrequest_id=helloid hello=world\n", strings.Trim(string(buf.Bytes()), "\x00"); want != have {
		t.Errorf("\nexpected %#v\n     got %#v", want, have)
	}
}
