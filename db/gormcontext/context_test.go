package gormcontext_test

import (
	"context"
	"testing"

	"github.com/go-midway/midway/db/gormcontext"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestWithDB(t *testing.T) {

	var ctx context.Context
	var dbSrc, dbGot *gorm.DB
	var err error

	// try empty context
	ctx = context.Background()
	dbGot = gormcontext.GetDB(ctx)
	if dbGot != nil {
		t.Errorf("unexpected value: %#v", dbGot)
	}

	// now with db
	dbSrc, err = gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	ctx = gormcontext.WithDB(ctx, dbSrc)
	dbGot = gormcontext.GetDB(ctx)
	if want, have := dbGot, dbSrc; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestWithNamedDB(t *testing.T) {

	var ctx context.Context
	var dbSrc1, dbSrc2, dbGot1, dbGot2 *gorm.DB
	var err error

	// try empty context
	ctx = context.Background()
	dbGot1 = gormcontext.GetNamedDB(ctx, "name 1")
	if dbGot1 != nil {
		t.Errorf("unexpected value: %#v", dbGot1)
	}
	dbGot2 = gormcontext.GetNamedDB(ctx, "name 2")
	if dbGot1 != nil {
		t.Errorf("unexpected value: %#v", dbGot2)
	}

	// now with db
	dbSrc1, err = gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	dbSrc2, err = gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	// put named db into context
	ctx = gormcontext.WithNamedDB(ctx, "name 1", dbSrc1)
	ctx = gormcontext.WithNamedDB(ctx, "name 2", dbSrc2)

	// test retrieving
	dbGot1 = gormcontext.GetNamedDB(ctx, "name 1")
	if want, have := dbGot1, dbSrc1; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
	dbGot2 = gormcontext.GetNamedDB(ctx, "name 2")
	if want, have := dbGot2, dbSrc2; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}
