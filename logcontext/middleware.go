package logcontext

import (
	"net/http"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-midway/midway"
)

// ApplyLogger logs access and also provide the kitlog context to inner
// http handler
func ApplyLogger(newlogger func() kitlog.Logger) midway.Middleware {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			reqID := r.Header.Get("X-Request-ID")
			logger := newlogger()
			logger = kitlog.With(
				logger,
				"request_id", reqID,
			)

			// access log
			logger.Log(
				"at", "info",
				"method", r.Method,
				"path", r.URL.Path,
				"protocol", r.URL.Scheme,
				"remote_addr", r.RemoteAddr,
			)

			inner.ServeHTTP(w, r.WithContext(WithLogger(r.Context(), logger)))
		})
	}
}
