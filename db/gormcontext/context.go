package gormcontext

import (
	"context"

	"github.com/jinzhu/gorm"
)

type contextKey int
type contextStrKey string

const (
	dbCtxKey contextKey = iota
)

// WithDB inserts a *gorm.DB into the context
func WithDB(parent context.Context, db *gorm.DB) context.Context {
	return context.WithValue(parent, dbCtxKey, db)
}

// GetDB returns a *gorm.DB or nil if the context have none
func GetDB(ctx context.Context) (db *gorm.DB) {
	db, _ = ctx.Value(dbCtxKey).(*gorm.DB)
	return
}

// WithNamedDB inserts a named *gorm.DB into the context, identified by string name
func WithNamedDB(parent context.Context, name string, db *gorm.DB) context.Context {
	return context.WithValue(parent, contextStrKey(name), db)
}

// GetNamedDB returns a named *gorm.DB or nil if the context have none
func GetNamedDB(ctx context.Context, name string) (db *gorm.DB) {
	db, _ = ctx.Value(contextStrKey(name)).(*gorm.DB)
	return
}
