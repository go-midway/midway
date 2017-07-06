package midway_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-midway/midway"
)

func TestHandleRequestID(t *testing.T) {
	catchID := false
	reqID := ""
	srv := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reqID = r.Header.Get("X-Request-ID"); reqID != "" {
			catchID = true
			fmt.Fprintf(w, "success: %s", reqID)
			return
		}
		fmt.Fprintf(w, "failed: got no request id")
	}))
	srv = midway.HandleRequestID(srv)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://foobar.com/hello/world", nil)
	srv.ServeHTTP(w, r)

	want := fmt.Sprintf("success: %s", reqID)
	if have := w.Body.String(); want != have {
		t.Errorf("expected: %#v, got: %#v", want, have)
	}
}
