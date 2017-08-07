package logcontext_test

import (
	"context"
	"os"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-midway/midway/logcontext"
)

func TestWithLogger(t *testing.T) {
	logger := kitlog.NewLogfmtLogger(os.Stdout)
	ctx := logcontext.WithLogger(context.Background(), logger)
	loggerOut := logcontext.GetLogger(ctx)
	if want, have := logger, loggerOut; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}
