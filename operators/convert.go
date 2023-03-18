package operators

import (
	"fmt"

	"github.com/avahidi/molly/types"
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
	Register("len", lenFunction)
}
