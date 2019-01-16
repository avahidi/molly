package util

import (
	"fmt"
	"reflect"
	"strings"
)

// Function represents a call to a golang function with the following format
//		func(param1, param2, ...) (ret1, ...)
type Function struct {
	name      string
	fun       reflect.Value
	ins, outs []reflect.Type
	variadic  bool
}

func NewFunction(name string, fun interface{}) (*Function, error) {
	t := reflect.TypeOf(fun)
	if t.Kind() != reflect.Func {
		return nil, fmt.Errorf("This interface is not a function: '%v'", fun)
	}

	v := reflect.ValueOf(fun)
	f := &Function{fun: v, name: name, variadic: t.IsVariadic()}

	for i := 0; i < t.NumOut(); i++ {
		f.outs = append(f.outs, t.Out(i))
	}
	for i := 0; i < t.NumIn(); i++ {
		f.ins = append(f.ins, t.In(i))
	}
	return f, nil
}

func (f Function) String() string {
	return f.name
}

func (f Function) Signature(input bool) []string {
	list := f.ins
	if !input {
		list = f.outs
	}

	ret := make([]string, len(list))
	for i, param := range list {
		ret[i] = strings.Replace(param.String(), "interface {}", "any", -1)
		if input && f.variadic && i == len(list)-1 {
			ret[i] = "..." + ret[i][2:] // []xx -> ...xx
		}
	}
	return ret
}

// Call does the actuall function call using golang reflection, after some
// format and type checking.
// This function will fail if the parameter and return format do not match.
// This function will recover panics and report them to RegisterFatalf()
func (f Function) Call(args []interface{}) ([]interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			RegisterFatalf("Call failed: %s", r)
		}
	}()

	var vart reflect.Type
	normalArgs := len(f.ins)

	if f.variadic {
		// check parameters in variadic functions
		if len(f.ins)-1 > len(args) {
			return nil, fmt.Errorf("Too few parameters in call to variadic to '%s(...)'", f.name)
		}
		vart = f.ins[normalArgs-1].Elem()
		normalArgs = normalArgs - 1
	} else {
		// check parameter count for normal functions
		if len(f.ins) != len(args) {
			return nil, fmt.Errorf("Too few parameters in call to %s(%v)", f.name, f.ins)
		}
	}
	rargs := make([]reflect.Value, len(args))
	for i, a := range args {
		v := reflect.ValueOf(a)
		t := reflect.TypeOf(a)

		// convert type if needed:
		newt := vart
		if i < normalArgs {
			newt = f.ins[i]
		}
		if t.ConvertibleTo(newt) {
			v = v.Convert(newt)
		} else {
			RegisterFatalf("In call to '%s': cannot convert parameter %d from '%v' to '%v'",
				f.name, i, t, newt)
		}
		rargs[i] = v
	}

	vals := f.fun.Call(rargs)

	// extract and convert return values
	rets := make([]interface{}, len(vals))
	for i, _ := range vals {
		rets[i] = vals[i].Convert(f.outs[i]).Interface()
	}
	return rets, nil
}

// FunctionDatabase is a database of cuntion with a given prototype
type FunctionDatabase struct {
	Functions map[string]*Function
	prototype *Function
}

func NewFunctionDatabase(prototype interface{}) (*FunctionDatabase, error) {
	p, err := NewFunction("", prototype)
	db := &FunctionDatabase{
		Functions: make(map[string]*Function),
		prototype: p,
	}
	return db, err
}

// Register registers a user function
func (db FunctionDatabase) Register(name string, fun interface{}) error {
	if _, found := db.Find(name); found {
		return fmt.Errorf("Attempted to register existing function %s", name)
	}

	f1, err := NewFunction(name, fun)
	if err != nil {
		return err
	}
	if len(f1.outs) != len(db.prototype.outs) || f1.outs[1] != db.prototype.outs[1] {
		return fmt.Errorf("Attempted to register function with incorrect return types: %v", f1.outs)
	}
	if len(f1.ins) < 1 || f1.ins[0] != db.prototype.ins[0] {
		return fmt.Errorf("Attempted to register function with incorrect parameter types: %v", f1.outs)
	}

	db.Functions[name] = f1
	return nil
}

// Find finds among registered functions
func (db FunctionDatabase) Find(name string) (*Function, bool) {
	f, found := db.Functions[name]
	return f, found
}
