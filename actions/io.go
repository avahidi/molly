package actions

import (
	"bitbucket.org/vahidi/molly/types"
)

func fileFunction(e *types.Env, prefix string) (string, error) {
	return e.Name(prefix, false)
}

func dirFunction(e *types.Env, prefix string) (string, error) {
	return e.Mkdir(prefix)
}

func init() {
	ActionRegister("file", fileFunction)
	ActionRegister("dir", dirFunction)
}
