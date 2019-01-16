package operators

import (
	"bitbucket.org/vahidi/molly/types"
)

func fileFunction(e *types.Env, prefix string) (string, error) {
	newname, _, err := e.New(prefix, false)
	return newname, err
}

func dirFunction(e *types.Env, prefix string) (string, error) {
	newname, _, err := e.Mkdir(prefix)
	return newname, err
}

func init() {
	Register("file", fileFunction)
	Register("dir", dirFunction)
}
