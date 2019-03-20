package operators

import (
	"bitbucket.org/vahidi/molly/types"
)

func fileFunction(e *types.Env, prefix string) (string, error) {
	data, err := e.New(prefix, false)
	if err != nil {
		return "", err
	}
	return data.Filename, err
}

func dirFunction(e *types.Env, prefix string) (string, error) {
	data, err := e.Mkdir(prefix)
	if err != nil {
		return "", err
	}
	return data.Filename, err
}

func init() {
	Register("file", fileFunction)
	Register("dir", dirFunction)
}
