package actions

import (
	"bitbucket.org/vahidi/molly/lib/types"
)

func fileFunction(e *types.Env, prefix string) (string, error) {
	return e.Name(prefix, false)
}

func dirFunction(e *types.Env, prefix string) (string, error) {
	return e.Mkdir(prefix)
}

func init() {
	types.FunctionRegister("file", fileFunction)
	types.FunctionRegister("dir", dirFunction)
}
