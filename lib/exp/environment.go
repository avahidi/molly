package exp

import (
	"bitbucket.org/vahidi/molly/lib/exp/prim"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
)

// encLookupSpecial checks for special variables that are from other sources
func encLookupSpecial(e *types.Env, id string) (interface{}, bool) {
	var val interface{}
	switch id {
	case "$filename":
		val = e.Input.Filename
	case "$filesize":
		val = e.Input.Filesize
	default:
		return nil, false
	}
	return val, true
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
		val, found = encLookupSpecial(e, id)
		if found {
			exp = NewValueExpression(prim.ValueToPrimitive(val))
		}
	}
	return exp, found, nil
}
