package exp

import (
	"bitbucket.org/vahidi/molly/lib/exp/prim"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util/logging"
)

// EnvLookup returns a variable either from current scope or
// the global registry
func EnvLookup(e *types.Env, id string) (types.Expression, bool, error) {
	exp, found := e.Scope.Get(id)

	// lazy evaluation in progress?
	if found && exp == nil {
		logging.Fatalf("Circular dependency in %s (%s)", id, e)
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

	if !found {
		var val interface{}
		val, found = e.Globals.Get(id)
		if found {
			exp = NewValueExpression(prim.ValueToPrimitive(val))
		}
	}
	return exp, found, nil
}
