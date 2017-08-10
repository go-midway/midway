package funconv

import (
	"fmt"
	"reflect"
	"testing"
)

func TestConvertError(t *testing.T) {
	var err error = &convertError{
		pos: 123,
		err: fmt.Errorf("some error"),
	}
	if want, have := "at 123, some error", err.Error(); want != have {
		t.Errorf("expected: %#v, got %#v", want, have)
	}
}

func TestMapConvertError(t *testing.T) {
	var err error = &mapConverterError{
		pos:  123,
		from: reflect.TypeOf("hello"),
		to:   reflect.TypeOf(0),
	}
	if want, have := "at 123, string cannot be converted to int", err.Error(); want != have {
		t.Errorf("expected: %#v, got %#v", want, have)
	}
}

func TestFindLastError(t *testing.T) {

	var firstError func() (error, int)
	var secondError func() (error, error, int)
	var thirdError func() (error, error, error, int)

	mapFuncRetTypes := func(fnType reflect.Type) (fnArgTypes []reflect.Type) {
		length := fnType.NumOut()
		fnArgTypes = make([]reflect.Type, length)
		for i := 0; i < length; i++ {
			fnArgTypes[i] = fnType.Out(i)
		}
		return
	}

	tests := []struct {
		name string
		fn   interface{}
		pos  int
	}{
		{
			name: "error at first",
			fn:   &firstError,
			pos:  0,
		},
		{
			name: "error at second",
			fn:   &secondError,
			pos:  1,
		},
		{
			name: "error at third",
			fn:   &thirdError,
			pos:  2,
		},
	}

	for _, test := range tests {
		types := mapFuncRetTypes(reflect.TypeOf(test.fn).Elem())
		if want, have := test.pos, findLastError(types); want != have {
			t.Errorf("at test %#v, expected %#v, got %#v",
				test.name, want, have)
		}
	}
}
