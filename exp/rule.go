package exp

import (
	"fmt"

	"github.com/avahidi/molly/exp/prim"
	"github.com/avahidi/molly/types"
	"github.com/avahidi/molly/util"
)

func RuleEval(rule *types.Rule, env *types.Env) (bool, error) {
	if env == nil {
		return false, fmt.Errorf("Rule operation requires a valid environment")
	}

	for _, n := range rule.Conditions {
		e, err := n.Eval(env)
		if err != nil {
			return false, err
		}
		ve, okay := e.(*ValueExpression)
		if !okay {
			return false, fmt.Errorf("condition is not a value expression: %t", e)
		}
		ne, okay1 := ve.Value.(*prim.Boolean)
		if !okay1 {
			return false, fmt.Errorf("condition is not a boolean expression: %t", e)
		}
		if !ne.Value {
			return false, nil
		}
	}

	// since this file evaluated to true, lets make sure we get
	// all its remaining assignments are computed
	for id, orgexp := range rule.Variables {
		if _, found := env.Scope.Get(id); !found {
			if exp, err := orgexp.Eval(env); err == nil {
				env.Scope.Set(id, exp)
			} else {
				// TODO: how do we report an error that happens AFTER
				// conditions have been met?
				return true, err
			}
		}
	}
	return true, nil
}

// RuleClose closes a newly read rule so it can be used for evaluation
func RuleClose(rule *types.Rule) {

	for id, v := range rule.Variables {
		rule.Variables[id] = Simplify(v)
	}

	for i, c := range rule.Conditions {
		rule.Conditions[i] = Simplify(c)
	}

	for i, a := range rule.Actions {
		rule.Actions[i].Action = Simplify(a.Action)
	}

	// at this point metadata for expressions don't point to rule as
	// its parent, this fixes that:
	var adjustMetadataParent visitor
	adjustMetadataParent = func(a types.Expression) visitor {
		var metadata *util.Register
		switch n := a.(type) {
		case *FunctionExpression:
			metadata = n.Metadata
		case *ExtractExpression:
			metadata = n.Metadata
		}
		if metadata != nil {
			metadata.SetParent(rule.Metadata)
		}
		return adjustMetadataParent
	}
	RuleVisitExpressions(rule, adjustMetadataParent)
}

// RuleVisitExpressions walks all expressions in a rule
func RuleVisitExpressions(rule *types.Rule, v visitor) {
	for _, a := range rule.Variables {
		walk(a, v)
	}
	for _, cc := range rule.Conditions {
		walk(cc, v)
	}
	for _, a := range rule.Actions {
		walk(a.Action, v)
	}
}
