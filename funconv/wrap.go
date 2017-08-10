package funconv

import (
	"fmt"
	"reflect"
)

// Converter converts one reflect.Value into another
type Converter func(reflect.Value) (reflect.Value, error)

// Convert the src to dest, according to the reflect.Value conversion
func (conv Converter) Convert(src interface{}) interface{} {
	srcValue := reflect.ValueOf(src)
	destValue, err := conv(srcValue)
	if err != nil {
		panic(err)
	}
	return destValue.Interface()
}

func makePipe(converters []Converter) func([]reflect.Value) ([]reflect.Value, error) {
	return func(in []reflect.Value) (out []reflect.Value, err error) {
		length := len(in)
		out = make([]reflect.Value, length)
		for i := range in {
			out[i], err = converters[i](in[i])
			if err != nil {
				err = &convertError{
					pos: i,
					err: err,
				}
				return
			}
		}
		return
	}
}

func directMap(src reflect.Value) (reflect.Value, error) {
	return src, nil
}

func convertToType(typ reflect.Type) Converter {
	return func(src reflect.Value) (reflect.Value, error) {
		return src.Convert(typ), nil
	}
}

func reverseConvertToItfType(typ reflect.Type) Converter {
	return func(src reflect.Value) (dest reflect.Value, err error) {

		srcInner := src
		if src.Kind() == reflect.Interface {
			srcInner = src.Elem()
		}

		if srcInner.Type() == typ {
			dest = srcInner
			return
		}
		if srcInner.Type().AssignableTo(typ) {
			dest = srcInner.Convert(typ)
			return
		}

		// TODO: return as error
		err = fmt.Errorf("%s cannot be converted to %s",
			srcInner.Type().String(), typ.String())
		return
	}
}

func findLastError(typs []reflect.Type) (pos int) {
	for pos = len(typs) - 1; pos >= 0; pos-- {
		if typs[pos].String() == "error" {
			return
		}
	}
	return -1
}

// ArgumentError is the runtime error in coverting
// arguments from wrapper function to the "real" function
type ArgumentError struct {
	pos int
	err error
}

func (err ArgumentError) Error() string {
	return fmt.Sprintf("argument %d, %s", err.pos+1, err.err.Error())
}

// RetVarError is the runtime error in coverting
// arguments from wrapper function to the "real" function
type RetVarError struct {
	pos int
	err error
}

func (err RetVarError) Error() string {
	return fmt.Sprintf("return variable %d, %s", err.pos+1, err.err.Error())
}

type convertError struct {
	pos int
	err error
}

func (err convertError) Error() string {
	return fmt.Sprintf("at %d, %s", err.pos, err.err.Error())
}

type mapConverterError struct {
	pos  int
	from reflect.Type
	to   reflect.Type
}

func (err mapConverterError) Error() string {
	return fmt.Sprintf("at %d, %s cannot be converted to %s",
		err.pos, err.from.String(), err.to.String())
}

// MapConverter maps converter slice for converting variable inTypes into outTypes
func MapConverter(inTypes []reflect.Type, outTypes []reflect.Type) (converters []Converter, err error) {
	length := len(inTypes)
	converters = make([]Converter, length)
	for i := 0; i < length; i++ {
		if inTypes[i] == outTypes[i] {
			//log.Printf("pos: %d, conveter: directMap", i)
			converters[i] = directMap
			continue
		}
		if inTypes[i].AssignableTo(outTypes[i]) {
			//log.Printf("pos: %d, conveter: convertToType", i)
			converters[i] = convertToType(outTypes[i])
			continue
		}
		if inTypes[i].Kind() == reflect.Interface && outTypes[i].Implements(inTypes[i]) {
			//log.Printf("pos: %d, conveter: reverseConvertToItfType", i)
			converters[i] = reverseConvertToItfType(outTypes[i])
			continue
		}
		err = &mapConverterError{
			pos:  i,
			from: inTypes[i],
			to:   outTypes[i],
		}
	}
	return
}

// WrapAs takes a funciton value (srcFunc), wrap it properly with type conversions
// then set it to function variable pointer (destFunc)
func WrapAs(srcFunc, destFunc interface{}) (err error) {

	//
	// validate arguments type
	//

	if srcFunc == nil {
		err = fmt.Errorf("srcFunc cannot be nil")
		return
	}
	if destFunc == nil {
		err = fmt.Errorf("destFunc cannot be nil")
		return
	}

	srcFuncType, destFuncType := reflect.TypeOf(srcFunc), reflect.TypeOf(destFunc)
	if srcFuncType.Kind() != reflect.Func {
		err = fmt.Errorf("srcFunc needs to be a function, got %T", srcFunc)
		return
	}
	if destFuncType.Kind() != reflect.Ptr {
		err = fmt.Errorf("destFunc needs to be a pointer of function variable, got %T", srcFunc)
		return
	}

	srcFuncVal := reflect.ValueOf(srcFunc)
	destFuncVal := reflect.ValueOf(destFunc).Elem()
	if destFuncVal.Kind() != reflect.Func {
		err = fmt.Errorf("destFunc needs to be a pointer of function variable, got %T", srcFunc)
		return
	}
	destFuncValType := destFuncVal.Type()

	if want, have := srcFuncType.NumIn(), destFuncValType.NumIn(); want != have {
		err = fmt.Errorf("argument mismatch, srcFunc(%d) != destFunc(%d)", want, have)
		return
	}
	if want, have := srcFuncType.NumOut(), destFuncValType.NumOut(); want != have {
		err = fmt.Errorf("return mismatch, srcFunc(%d) != destFunc(%d)", want, have)
		return
	}

	//
	// variadic check
	//
	if srcFuncType.IsVariadic() && !destFuncValType.IsVariadic() {
		err = fmt.Errorf("srcFunc is variadic function while destFunc is not")
		return
	}
	if !srcFuncType.IsVariadic() && destFuncValType.IsVariadic() {
		err = fmt.Errorf("destFunc is variadic function while srcFunc is not")
		return
	}

	//
	// compose functions for arguments and return variables conversions
	//
	numIn, numOut := srcFuncType.NumIn(), srcFuncType.NumOut()

	// generate function input converters
	outerInTypes := make([]reflect.Type, numIn)
	innerInTypes := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		innerInTypes[i], outerInTypes[i] = srcFuncType.In(i), destFuncValType.In(i)
	}
	inConverters, err := MapConverter(outerInTypes, innerInTypes)
	if err != nil {
		innerErr := err.(*mapConverterError)
		err = fmt.Errorf(
			"argument %d, %s cannot be converted %s",
			innerErr.pos+1, innerErr.from.String(), innerErr.to.String())
		return
	}
	funcIn := makePipe(inConverters)

	// generate function output converters
	outerOutTypes := make([]reflect.Type, numOut)
	innerOutTypes := make([]reflect.Type, numOut)
	for i := 0; i < numOut; i++ {
		innerOutTypes[i], outerOutTypes[i] = srcFuncType.Out(i), destFuncValType.Out(i)
	}
	outConverters, err := MapConverter(innerOutTypes, outerOutTypes)
	if err != nil {
		innerErr := err.(*mapConverterError)
		err = fmt.Errorf(
			"return variable %d, %s cannot be converted %s",
			innerErr.pos+1, innerErr.from.String(), innerErr.to.String())
		return
	}
	funcOut := makePipe(outConverters)

	// default error handling
	handleError := func(err error, out []reflect.Value) []reflect.Value {
		panic(err)
	}

	// find the "error" return variable in out
	// if any, handle the error with the last error parameter
	if pos := findLastError(outerOutTypes); pos >= 0 {
		handleError = func(err error, out []reflect.Value) []reflect.Value {
			// initialize output variables properly, if not already
			for i := range outerOutTypes {
				if i != pos && !out[i].IsValid() {
					out[i] = reflect.Zero(outerOutTypes[i])
				}
			}

			// set the conversion error to the output variable
			out[pos] = reflect.ValueOf(err).Convert(outerOutTypes[pos])
			return out
		}
	}

	// compose the wrapped function
	resultFunc := reflect.MakeFunc(destFuncValType, func(in []reflect.Value) (out []reflect.Value) {
		// convert input arguments
		innerIn, err := funcIn(in)
		if err != nil {
			// initialize the output array
			innerErr := err.(*convertError)
			return handleError(
				&ArgumentError{
					pos: innerErr.pos,
					err: innerErr.err,
				},
				make([]reflect.Value, numOut),
			)
		}

		// call srcFunc function
		innerOut := srcFuncVal.Call(innerIn)

		// convert return variables
		out, err = funcOut(innerOut)
		if err != nil {
			innerErr := err.(*convertError)
			return handleError(
				&RetVarError{
					pos: innerErr.pos,
					err: innerErr.err,
				},
				out,
			)
		}

		return
	})
	// set the resultFunc to the pointer of destFunc
	destFuncVal.Set(resultFunc)
	return
}
