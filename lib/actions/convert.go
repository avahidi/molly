package actions

import (
	"fmt"
	"strconv"

	"bitbucket.org/vahidi/molly/lib/types"
)

func lenFunction(e *types.Env, item interface{}) (interface{}, error) {
	switch n := item.(type) {
	case string:
		return len(n), nil
	default:
		return nil, fmt.Errorf("Cannot decide length of item '%v'", item)
	}
}

func toIntFunction(e *types.Env, str string) (interface{}, error) {
	return strconv.ParseInt(str, 10, 64)
}

func init() {
	types.FunctionRegister("len", lenFunction)
	types.FunctionRegister("toInt", toIntFunction)
}
