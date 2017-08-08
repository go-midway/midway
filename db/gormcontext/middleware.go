package gormcontext

import (
	"net/http"

	"github.com/go-midway/midway"
	"github.com/jinzhu/gorm"
)

// ApplyDB puts a *gorm.DB into the context for inner handler
func ApplyDB(db *gorm.DB) midway.Middleware {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inner.ServeHTTP(w, r.WithContext(WithDB(r.Context(), db)))
		})
	}
}

// ApplyNamedDB puts a *gorm.DB into the context for inner handler
func ApplyNamedDB(db *gorm.DB, name string) midway.Middleware {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inner.ServeHTTP(w, r.WithContext(WithNamedDB(r.Context(), name, db)))
		})
	}
}
