package types

import (
	"fmt"
	"log"

	"github.com/avahidi/molly/util"
)

// operatorPrototype is a template for molly operators
func operatorPrototype(e *Env, args ...interface{}) (interface{}, error) {
	return nil, nil
}

var operators *util.FunctionDatabase

// OperatorRegister registers a new operator in molly
func OperatorRegister(name string, fun interface{}) error {
	return operators.Register(name, fun)
}

// OperatorFind finds among registered functions
func OperatorFind(name string) (*util.Function, bool) {
	return operators.Find(name)
}

// OperatorHelp print information about all known operators
func OperatorHelp() {
	fmt.Printf("Available operators are:\n")
	for _, v := range operators.Functions {
		ins, outs := v.Signature(true), v.Signature(false)
		fmt.Printf("\t%-12s (%s) -> %s\n", v, ins[1:], outs[0])
	}
}

func init() {
	var err error
	operators, err = util.NewFunctionDatabase(operatorPrototype)
	if err != nil {
		log.Panic(err)
	}
}
