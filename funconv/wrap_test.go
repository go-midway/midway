package funconv_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/go-midway/midway/funconv"
)

type stringer int

func (s stringer) String() string {
	return "some stringer"
}

func TestMapConverter_simpleFuncIn(t *testing.T) {

	var strIn func(string)
	var itfceIn func(interface{})

	mapFuncArgTypes := func(fnType reflect.Type) (fnArgTypes []reflect.Type) {
		length := fnType.NumIn()
		fnArgTypes = make([]reflect.Type, length)
		for i := 0; i < length; i++ {
			fnArgTypes[i] = fnType.In(i)
		}
		return
	}

	tests := []struct {
		name    string
		inFunc  interface{}
		outFunc interface{}
		toConv  []interface{}
		want    []interface{}
	}{
		{
			name:    "direct map",
			inFunc:  func(string) {},
			outFunc: &strIn,
			toConv:  []interface{}{"hello"},
			want:    []interface{}{"hello"},
		},
		{
			name:    "string to interface{}",
			inFunc:  func(string) {},
			outFunc: &itfceIn,
			toConv:  []interface{}{"hello"},
			want:    []interface{}{"hello"},
		},
	}
	for i, test := range tests {
		inTypes := mapFuncArgTypes(reflect.TypeOf(test.inFunc))
		outTypes := mapFuncArgTypes(reflect.ValueOf(test.outFunc).Elem().Type())
		conv, err := funconv.MapConverter(inTypes, outTypes)
		if err != nil {
			t.Errorf("unexpected error in test %#v: %s",
				test.name, err.Error())
			continue
		}

		// test each toConv in their position of argument
		// i.e. toConv[1] = argument[1]
		//      want[1] = expectedConvertedArgument[1]
		for j := range test.toConv {

			// find converter
			if conv[j] == nil {
				t.Errorf("test %d.%d: converter is nil", i, j)
				continue
			}

			// test convert with mapped converter
			out := conv[j].Convert(test.toConv[j])
			if want, have := test.want[j], out; want != have {
				t.Errorf("test %d.%d: expected %#v, got %#v",
					i, j, want, have)
			}
		}
	}
}

func TestMapConverter_interfaceFuncIn(t *testing.T) {

	var stringerIn func(stringer)
	var fmtStringerIn func(fmt.Stringer)

	mapFuncArgTypes := func(fnType reflect.Type) (fnArgTypes []reflect.Type) {
		length := fnType.NumIn()
		fnArgTypes = make([]reflect.Type, length)
		for i := 0; i < length; i++ {
			fnArgTypes[i] = fnType.In(i)
		}
		return
	}

	tests := []struct {
		name    string
		inFunc  interface{}
		outFunc interface{}
		toConv  []interface{}
	}{
		{
			name:    "stringer to stringer",
			inFunc:  func(stringer) {},
			outFunc: &stringerIn,
			toConv:  []interface{}{stringer(1)},
		},
		{
			name:    "stringer to fmt.Stringer",
			inFunc:  func(stringer) {},
			outFunc: &fmtStringerIn,
			toConv:  []interface{}{stringer(1)},
		},
		{
			name:    "fmt.Stringer to stringer",
			inFunc:  func(fmt.Stringer) {},
			outFunc: &stringerIn,
			toConv:  []interface{}{stringer(1)},
		},
	}
	for i, test := range tests {
		inTypes := mapFuncArgTypes(reflect.TypeOf(test.inFunc))
		outTypes := mapFuncArgTypes(reflect.ValueOf(test.outFunc).Elem().Type())
		conv, err := funconv.MapConverter(inTypes, outTypes)
		if err != nil {
			t.Errorf("unexpected error in test %#v: %s",
				test.name, err.Error())
			continue
		}

		j := 0

		// find converter
		if conv[j] == nil {
			t.Errorf("test %d: converter is nil", i)
			continue
		}

		// test convert with mapped converter
		out := conv[j].Convert(test.toConv[j])
		if out == nil {
			t.Errorf("expect out to be not nil")
		}
	}
}

func TestWrap_mismatchArg(t *testing.T) {
	var err error

	// srcFunc and destFunc
	lengthFunc := func(ctx context.Context, name string) (length int, err error) {
		length = len(name)
		return
	}
	var mismatchArg1 func(req interface{}) (resp interface{}, err error)
	var mismatchArg3 func(req1, req2, req3 interface{}) (resp interface{}, err error)

	// trigger error with unmatch argument number
	err = funconv.WrapAs(lengthFunc, &mismatchArg1)
	if err == nil {
		t.Errorf("expecting error, got nil")
	} else if want, have := fmt.Sprintf("argument mismatch, srcFunc(%d) != destFunc(%d)", 2, 1), err.Error(); want != have {
		t.Errorf("exptected %#v, got %#v", want, have)
	}

	// trigger error with unmatch argument number
	err = funconv.WrapAs(lengthFunc, &mismatchArg3)
	if err == nil {
		t.Errorf("expecting error, got nil")
	} else if want, have := fmt.Sprintf("argument mismatch, srcFunc(%d) != destFunc(%d)", 2, 3), err.Error(); want != have {
		t.Errorf("exptected %#v, got %#v", want, have)
	}
}

func TestWrap_mismatchReturn(t *testing.T) {
	var err error

	// srcFunc and destFunc
	lengthFunc := func(ctx context.Context, name string) (length int, err error) {
		length = len(name)
		return
	}
	var mismatchRet1 func(ctx context.Context, req interface{}) (resp interface{})
	var mismatchRet3 func(ctx context.Context, req interface{}) (resp1, resp2 interface{}, err error)

	// trigger error with unmatch argument number
	err = funconv.WrapAs(lengthFunc, &mismatchRet1)
	if err == nil {
		t.Errorf("expecting error, got nil")
	} else if want, have := fmt.Sprintf("return mismatch, srcFunc(%d) != destFunc(%d)", 2, 1), err.Error(); want != have {
		t.Errorf("exptected %#v, got %#v", want, have)
	}

	// trigger error with unmatch argument number
	err = funconv.WrapAs(lengthFunc, &mismatchRet3)
	if err == nil {
		t.Errorf("expecting error, got nil")
	} else if want, have := fmt.Sprintf("return mismatch, srcFunc(%d) != destFunc(%d)", 2, 3), err.Error(); want != have {
		t.Errorf("exptected %#v, got %#v", want, have)
	}
}

func TestWrap_variadicCheck(t *testing.T) {

	var err error

	// srcFunc and destFunc
	lengthString := func(ctx context.Context, name string) (length int, err error) {
		length = len(name)
		return
	}
	lengthArray := func(ctx context.Context, name ...string) (length int, err error) {
		length = len(name)
		return
	}
	var destFunc func(req1, req2 string) (resp interface{}, err error)
	var destFuncVariadic func(req1 string, req2 ...string) (resp interface{}, err error)

	err = funconv.WrapAs(lengthString, &destFuncVariadic)
	if err == nil {
		t.Errorf("expecting error, got nil")
	} else if want, have := "destFunc is variadic function while srcFunc is not", err.Error(); want != have {
		t.Errorf("exptected %#v, got %#v", want, have)
	}

	err = funconv.WrapAs(lengthArray, &destFunc)
	if err == nil {
		t.Errorf("expecting error, got nil")
	} else if want, have := "srcFunc is variadic function while destFunc is not", err.Error(); want != have {
		t.Errorf("exptected %#v, got %#v", want, have)
	}

}

func TestWrap_passthrough(t *testing.T) {
	var err error
	// srcFunc and destFunc
	lengthFunc := func(name string) (length int) {
		length = len(name)
		return
	}
	var endpoint func(name string) (length int)

	err = funconv.WrapAs(lengthFunc, &endpoint)
	if err != nil {
		t.Errorf("unexpected error: %#v", err.Error())
	}

	if want, have := 5, endpoint("hello"); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestWrap_interfaceIn(t *testing.T) {
	var err error
	// srcFunc and destFunc
	lengthFunc := func(name string) (length int) {
		length = len(name)
		return
	}
	var endpoint func(name interface{}) (length int)

	err = funconv.WrapAs(lengthFunc, &endpoint)
	if err != nil {
		t.Errorf("unexpected error: %#v", err.Error())
	}

	if want, have := 5, endpoint("hello"); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestWrap_errorOutVal(t *testing.T) {
	var err error
	// srcFunc and destFunc
	dummyFunc := func(name string) (err error) {
		return nil
	}
	var endpoint func(name string) (err error)

	err = funconv.WrapAs(dummyFunc, &endpoint)
	if err != nil {
		t.Errorf("unexpected error: %#v", err.Error())
	}

	err = endpoint("hello")
	if err != nil {
		t.Errorf("unexpected error: %#v", err.Error())
	}
}

func TestWrap_errorOutNil(t *testing.T) {
	var err error
	// srcFunc and destFunc
	dummyFunc := func(name string) (err error) {
		return fmt.Errorf("some error")
	}
	var endpoint func(name string) (err error)

	err = funconv.WrapAs(dummyFunc, &endpoint)
	if err != nil {
		t.Errorf("unexpected error: %#v", err.Error())
	}

	err = endpoint("hello")
	if err == nil {
		t.Errorf("expected error but got nil")
		return
	}

	if want, have := "some error", err.Error(); want != have {
		t.Errorf("expected %#v but got %#v", want, have)
	}
}

func TestWrap_argTypeMismatch(t *testing.T) {
	var err error

	innerFunc1 := func(string) int {
		return 0
	}
	var endpoint func(int) int
	err = funconv.WrapAs(innerFunc1, &endpoint)
	if err == nil {
		t.Errorf("expected error but got nil")
		return
	}
	expectedMsg := "argument 1, int cannot be converted string"
	if want, have := expectedMsg, err.Error(); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestWrap_retTypeMismatch(t *testing.T) {
	var err error

	innerFunc1 := func(string) int {
		return 0
	}
	var endpoint func(string) string
	err = funconv.WrapAs(innerFunc1, &endpoint)
	if err == nil {
		t.Errorf("expected error but got nil")
		return
	}
	expectedMsg := "return variable 1, int cannot be converted string"
	if want, have := expectedMsg, err.Error(); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestWrap_endpointInputCastingError(t *testing.T) {
	var err error

	// srcFunc and destFunc
	lengthFunc := func(ctx context.Context, name string) (length int, err error) {
		length = len(name)
		return
	}
	var endpoint1 func(ctx context.Context, req int) (resp interface{}, err error)
	var endpoint2 func(ctx context.Context, req interface{}) (resp interface{}, err error)

	// test1: make error
	err = funconv.WrapAs(lengthFunc, &endpoint1)
	if err == nil {
		t.Errorf("expected error, got nil")
	} else if want, have := "argument 2, int cannot be converted string", err.Error(); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
		return
	}

	// test2: convert error
	err = funconv.WrapAs(lengthFunc, &endpoint2)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	_, err = endpoint2(nil, 123)
	if err == nil {
		t.Errorf("expected error, got nil")
	} else if want, have := "argument 2, int cannot be converted to string", err.Error(); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestWrap_endpointOutputCastingError(t *testing.T) {

	var err error

	// srcFunc and destFunc
	faultyFunc := func(ctx context.Context, name string) (length interface{}, err error) {
		return "some length", nil
	}
	var endpoint func(ctx context.Context, name string) (length int, err error)

	err = funconv.WrapAs(faultyFunc, &endpoint)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	_, err = endpoint(nil, "hello")
	if err == nil {
		t.Errorf("expected error, got nil")
	} else if want, have := "return variable 1, string cannot be converted to int", err.Error(); want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}

func TestWrap_endpointPanic(t *testing.T) {
	var err error

	// recover from panic message
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				t.Errorf("r is not an error, is %#v", err)
			} else if want, have := "return variable 1, string cannot be converted to int", err.Error(); want != have {
				t.Errorf("expected %#v, got %#v", want, have)
			}
		}
	}()

	// srcFunc and destFunc
	faultyFunc := func(ctx context.Context, name string) (length interface{}) {
		return "some length"
	}
	var endpoint func(ctx context.Context, name string) (length int)

	err = funconv.WrapAs(faultyFunc, &endpoint)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	endpoint(nil, "hello")
}

func TestWrap_endpoint(t *testing.T) {

	var err error

	// srcFunc and destFunc
	lengthFunc := func(ctx context.Context, name string) (length int, err error) {
		length = len(name)
		return
	}
	var endpoint func(ctx context.Context, req interface{}) (resp interface{}, err error)

	// trigger error with nil srcFunc
	err = funconv.WrapAs(nil, nil)
	if err == nil {
		t.Errorf("expecting error, got nil")
	} else if want, have := "srcFunc cannot be nil", err.Error(); want != have {
		t.Errorf("exptected %#v, got %#v", want, have)
	}

	// trigger error with nil destFunc
	err = funconv.WrapAs(lengthFunc, nil)
	if err == nil {
		t.Errorf("expecting error, got nil")
	} else if want, have := "destFunc cannot be nil", err.Error(); want != have {
		t.Errorf("exptected %#v, got %#v", want, have)
	}

	// trigger error with non-func srcFunc
	err = funconv.WrapAs("some stupid", &endpoint)
	if err == nil {
		t.Errorf("expecting error, got nil")
	} else if strings.HasPrefix("srcFunc needs to be a function, got ", err.Error()) {
		t.Errorf("unexpected error string pattern: %s", err.Error())
	}

	// trigger error with non-func-pointer distFunc
	err = funconv.WrapAs(lengthFunc, endpoint)
	if err == nil {
		t.Errorf("expecting error, got nil")
	} else if strings.HasPrefix("destFunc needs to be a pointer of function variable, got ", err.Error()) {
		t.Errorf("unexpected error string pattern: %s", err.Error())
	}

	// trigger error with non-func-pointer distFunc
	var randVar int
	err = funconv.WrapAs(lengthFunc, &randVar)
	if err == nil {
		t.Errorf("expecting error, got nil")
	} else if strings.HasPrefix("destFunc needs to be a pointer of function variable, got ", err.Error()) {
		t.Errorf("unexpected error string pattern: %s", err.Error())
	}

	// make the function
	err = funconv.WrapAs(lengthFunc, &endpoint)
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
	if endpoint == nil {
		t.Errorf("endpoint is still nil")
		return
	}
	resp, err := endpoint(nil, "hello world")
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
	if want, have := 11, resp; want != have {
		t.Errorf("expected %#v, got %#v", want, have)
	}
}
