package actions

import (
	"bitbucket.org/vahidi/molly/lib/types"
	"fmt"
)

func lenFunction(e *types.Env, item interface{}) (int, error) {
	switch n := item.(type) {
	case string:
		return len(n), nil
	default:
		return 0, fmt.Errorf("Cannot decide length of item '%v'", item)
	}
}

func init() {
	types.FunctionRegister("len", lenFunction)
}
