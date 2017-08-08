package gormcontext_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-midway/midway/db/gormcontext"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestApplyDB(t *testing.T) {
	var dbSrc *gorm.DB
	var err error

	dbSrc, err = gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	handler := gormcontext.ApplyDB(dbSrc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dbGot := gormcontext.GetDB(r.Context())
		if want, have := dbSrc, dbGot; want != have {
			t.Errorf("expected %#v, got %#v", want, have)
			fmt.Fprintf(w, "failed")
			return
		}
		fmt.Fprintf(w, "success")
	}))

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/foobar", nil)
	handler.ServeHTTP(w, r)
}

func TestApplyNamedDB(t *testing.T) {
	var dbSrc *gorm.DB
	var err error

	dbSrc, err = gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	handler := gormcontext.ApplyNamedDB(dbSrc, "name 1")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dbGot := gormcontext.GetNamedDB(r.Context(), "name 1")
		if want, have := dbSrc, dbGot; want != have {
			t.Errorf("expected %#v, got %#v", want, have)
			fmt.Fprintf(w, "failed")
			return
		}
		fmt.Fprintf(w, "success")
	}))

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/foobar", nil)
	handler.ServeHTTP(w, r)
}
