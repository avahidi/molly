package exp

import (
	"strings"

	"bitbucket.org/vahidi/molly/exp/prim"
	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
)

// envLookupSpecial checks for special variables that are from other sources
func envLookupSpecial(e *types.Env, id string) (interface{}, bool) {
	if strings.HasPrefix(id, "$") {
		return e.Input.Get(id[1:])
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

	// maybe a special variable?
	if !found {
		var val interface{}
		val, found = envLookupSpecial(e, id)
		if found {
			exp = NewValueExpression(prim.ValueToPrimitive(val))
		}
	}
	return exp, found, nil
}
