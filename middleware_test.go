package midway_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-midway/midway"
)

func TestChain(t *testing.T) {
	i, pm1, pm2 := 0, 0, 0
	m1 := func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			i++
			pm1 = i
			inner.ServeHTTP(w, r)
		})
	}
	m2 := func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			i++
			pm2 = i
			inner.ServeHTTP(w, r)
		})
	}
	srv := midway.Chain(m1, m2)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// do nothing
	}))

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://foobar.com/hello/world", nil)
	srv.ServeHTTP(w, r)
	if want, have := 1, pm1; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	if want, have := 2, pm2; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}

	srv.ServeHTTP(w, r)
	if want, have := 3, pm1; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	if want, have := 4, pm2; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}
