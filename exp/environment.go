package exp

import (
	"strings"

	"github.com/avahidi/molly/exp/prim"
	"github.com/avahidi/molly/types"
	"github.com/avahidi/molly/util"
)

// envLookupEnvironment checks for environment variables instead of rules
func envLookupEnvironment(e *types.Env, id string) (interface{}, bool) {
	if strings.HasPrefix(id, "$") {
		if val, found := e.Current.Get(id[1:]); found {
			return val, true
		}
		types.FileDataGetHelp()
	}
	return nil, false
}

// EnvLookup returns a variable either from current scope or
// the global registry
func EnvLookup(e *types.Env, id string) (types.Expression, bool, error) {
	exp, found := e.Scope.Get(id)

	// lazy evaluation in progress?
	if found && exp == nil {
		util.RegisterFatalf("Circular dependency in %s (%s)", id, e)
	}
	// attempt resolve lazy evaluation
	if !found {
		if exp, found = e.Scope.Rule.Variables[id]; found {
			e.Scope.Set(id, nil) // show that we are working on it...
			var err error
			exp, err = exp.Eval(e)
			if err != nil {
				return nil, true, err
			}
			e.Scope.Set(id, exp)
		}
	}

	// maybe a environment variable?
	if !found {
		var val interface{}
		val, found = envLookupEnvironment(e, id)
		if found {
			exp = NewValueExpression(prim.ValueToPrimitive(val))
		}
	}
	return exp, found, nil
}
