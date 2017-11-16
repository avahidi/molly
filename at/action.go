package at

import (
	"fmt"
	"reflect"
)

type ActionFunction struct {
	name string
	fun  reflect.Value
}

func (af ActionFunction) String() string {
	return af.name
}
func (af ActionFunction) Call(env *Env, args []interface{}) (r interface{}, ecall error, efatal error) {
	defer func() {
		if r := recover(); r != nil {
			efatal = fmt.Errorf("Call failed: %s", r)
		}
	}()

	// convert input to reflection
	rargs := make([]reflect.Value, 1+len(args))
	rargs[0] = reflect.ValueOf(env)
	for i, a := range args {
		rargs[i+1] = reflect.ValueOf(a)
	}

	ret := af.fun.Call(rargs)

	// extract and convert return values
	if len(ret) > 2 {
		return nil, nil, fmt.Errorf("Unexpected return type: %T", ret)
	}
	var r0 interface{} = false // false is a placeholder to avoid nil returns
	if len(ret) > 0 {
		if !ret[0].IsNil() {
			r0 = ret[0].Interface()
		}
	}

	var r1 error = nil
	if len(ret) > 1 {
		if !ret[1].IsNil() {
			r1 = ret[1].Interface().(error)
		}
	}

	return r0, r1, nil
}

var actionRegister = make(map[string]*ActionFunction)

func RegisterActionFunction(name string, fun interface{}) {
	actionRegister[name] = &ActionFunction{
		fun:  reflect.ValueOf(fun),
		name: name,
	}
}

func FindActionFunction(name string) (*ActionFunction, bool) {
	f, found := actionRegister[name]
	return f, found
}
