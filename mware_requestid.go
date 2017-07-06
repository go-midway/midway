package midway

import (
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// HandleRequestID is a middleware to apply X-Request-ID to request header
// if it is not set
func HandleRequestID(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := r.Header.Get("X-Request-ID"); id == "" {
			r.Header.Set("X-Request-ID", randStringRunes(20))
		}
		inner.ServeHTTP(w, r)
	})
}
