package logcontext

import (
	"context"
	"net/http"
	"os"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-midway/midway"
)

type contextKey int

const (
	logCtxKey contextKey = iota
	errLogCtxKey
)

// WithLogger stores a go-kit log *Context to a context.Context
func WithLogger(parent context.Context, logCtx kitlog.Logger) context.Context {
	return context.WithValue(parent, logCtxKey, logCtx)
}

// GetLogger get the go-kit log *Context from the context.Context
func GetLogger(ctx context.Context) (logCtx kitlog.Logger) {
	logCtx, _ = ctx.Value(logCtxKey).(kitlog.Logger)
	if logCtx == nil {
		logCtx = kitlog.NewLogfmtLogger(os.Stdout)
	}
	return
}

// WithErrLogger stores a go-kit log *Context to a context.Context
func WithErrLogger(parent context.Context, logCtx kitlog.Logger) context.Context {
	return context.WithValue(parent, errLogCtxKey, logCtx)
}

// GetErrLogger get the go-kit log *Context from the context.Context
func GetErrLogger(ctx context.Context) (logCtx kitlog.Logger) {
	logCtx, _ = ctx.Value(errLogCtxKey).(kitlog.Logger)
	if logCtx == nil {
		logCtx = kitlog.NewLogfmtLogger(os.Stderr)
	}
	return
}

// ComplexLogger returns a logger that does
// log differently for different log level
type ComplexLogger interface {
	Log(keyvals ...interface{}) error
	Error(keyvals ...interface{}) error
}

type complexLogger struct {
	info kitlog.Logger
	err  kitlog.Logger
}

func (loggers *complexLogger) Log(keyvals ...interface{}) error {
	return loggers.info.Log(keyvals...)
}

func (loggers *complexLogger) Error(keyvals ...interface{}) error {
	return loggers.err.Log(keyvals...)
}

// GetComplexLogger gets a complex logger from context
func GetComplexLogger(ctx context.Context) (logger ComplexLogger) {
	return &complexLogger{
		info: GetLogger(ctx),
		err:  GetErrLogger(ctx),
	}
}

// ProvideLoggers provides info and error loggers to inner handler
func ProvideLoggers(info, err kitlog.Logger) midway.Middleware {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := WithErrLogger(WithLogger(r.Context(), info), err)
			inner.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
