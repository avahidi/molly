package types

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"bitbucket.org/vahidi/molly/lib/util"
)

// Function represents a call to a golang function with the following format
//		func(e Env, param1, param2, ...) (ret1, error)
// These can be called from the rules, and user can add own functions
// using #FunctionRegister
type Function struct {
	name      string
	fun       reflect.Value
	ins, outs []reflect.Type
	variadic  bool
}

func newFunction(name string, fun interface{}) *Function {
	v := reflect.ValueOf(fun)
	t := reflect.TypeOf(fun)
	f := &Function{fun: v, name: name, variadic: t.IsVariadic()}

	if t.Kind() != reflect.Func {
		util.RegisterFatalf("This interface is not a function: '%v'", fun)
	}

	for i := 0; i < t.NumOut(); i++ {
		f.outs = append(f.outs, t.Out(i))
	}
	for i := 0; i < t.NumIn(); i++ {
		f.ins = append(f.ins, t.In(i))
	}
	return f
}

func (f Function) String() string {
	return f.name
}

// Signature returns function signature in human readable-form
// and excluding molly internal elements
func (f Function) Signature() (string, string, string) {
	buf := bytes.Buffer{}
	for i, ins := range f.ins[1:] {
		if i > 0 {
			buf.WriteString(", ")
		}
		str := strings.Replace(ins.String(), "interface {}", "any", -1)
		if f.variadic && i == len(f.ins)-2 {
			str = "..." + str[2:] // []xx -> ...xx
		}
		buf.WriteString(str)
	}
	out := strings.Replace(f.outs[0].String(), "interface {}", "any", -1)
	return f.name, buf.String(), out
}

// Call does the actuall function call using golang reflection, after some
// format and type checking.
// This function will fatally fail if the parameter and return format do not match.
func (f Function) Call(env *Env, args []interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			util.RegisterFatalf("Call failed: %s", r)
		}
	}()

	var vart reflect.Type
	normalArgs := len(f.ins)

	if f.variadic {
		// check parameters in variadic functions
		if len(f.ins)-2 > len(args) {
			util.RegisterFatalf("Too few parameters in call to variadic to '%s(...)'", f.name)
		}
		vart = f.ins[normalArgs-1].Elem()
		normalArgs = normalArgs - 1
	} else {
		// check parameter count for normal functions
		if len(f.ins) != len(args)+1 {
			util.RegisterFatalf("Too few parameters in call to %s(%v)", f.name, f.ins[1:])
		}
	}
	rargs := make([]reflect.Value, 1+len(args))
	rargs[0] = reflect.ValueOf(env)
	for i, a := range args {
		v := reflect.ValueOf(a)
		t := reflect.TypeOf(a)

		// convert type if needed:
		newt := vart
		if i+1 < normalArgs {
			newt = f.ins[i+1]
		}
		if t.ConvertibleTo(newt) {
			v = v.Convert(newt)
		} else {
			util.RegisterFatalf("In call to '%s': cannot convert parameter %d from '%v' to '%v'",
				f.name, i+1, t, newt)
		}
		rargs[i+1] = v
	}
	ret := f.fun.Call(rargs)

	// extract and convert return values
	r0 := ret[0].Convert(f.outs[0]).Interface()
	r1 := ret[1].Convert(f.outs[1]).Interface()
	if r1 == nil {
		return r0, nil
	}
	return r0, r1.(error)
}

// the prototype function is used for comparison with registered functions
var prototype *Function = newFunction("dummy", func(e *Env, args ...interface{}) (interface{}, error) { return nil, nil })

var actionRegister = make(map[string]*Function)

// FunctionRegister registers a user function
func FunctionRegister(name string, fun interface{}) error {
	if _, found := FunctionFind(name); found {
		return fmt.Errorf("Attempted to register existing function %s", name)
	}

	f1 := newFunction(name, fun)
	if len(f1.outs) != len(prototype.outs) || f1.outs[1] != prototype.outs[1] {
		return fmt.Errorf("Attempted to register function with incorrect return types: %v", f1.outs)
	}
	if len(f1.ins) < 1 || f1.ins[0] != prototype.ins[0] {
		return fmt.Errorf("Attempted to register function with incorrect parameter types: %v", f1.outs)
	}

	actionRegister[name] = f1
	return nil
}

// FunctionFind finds among registered functions
func FunctionFind(name string) (*Function, bool) {
	f, found := actionRegister[name]
	return f, found
}

// FunctionHelp prints help text including signature for all registred functions
func FunctionHelp() {
	fmt.Printf("Available functions are:\n")
	for _, v := range actionRegister {
		name, in, out := v.Signature()
		fmt.Printf("\t%-12s (%s) -> %s\n", name, in, out)
	}
}
