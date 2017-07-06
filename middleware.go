package midway

import "net/http"

// Middleware defines function signature of a middleware
type Middleware func(http.Handler) http.Handler

// Chain chains HTTPMiddleware to form a single middleware
func Chain(mwares ...Middleware) Middleware {
	return func(inner http.Handler) http.Handler {
		for i := len(mwares) - 1; i >= 0; i-- {
			inner = mwares[i](inner)
		}
		return inner
	}
}
