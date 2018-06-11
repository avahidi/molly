package scan

import (
	"fmt"
	"os"

	"bitbucket.org/vahidi/molly/exp"
	"bitbucket.org/vahidi/molly/types"
)

// executeRuleActions runs actions while taking into account their failure modes
func executeRuleActions(rule *types.Rule, env *types.Env) []error {
	var errors []error
	for _, a := range rule.Actions {
		// make sure all actions start from the beginning of the file
		env.Input.Seek(0, os.SEEK_SET)
		_, err := a.Action.Eval(env)
		if err == nil {
			if a.Mode == types.ActionModeExit {
				break
			}
		} else {
			if a.Mode == types.ActionModeNormal {
				err := fmt.Errorf("[action failure] %.100q...", err)
				errors = append(errors, err)
				break
			}
			err := fmt.Errorf("[action failure (ignored)] %.100q...", err)
			errors = append(errors, err)
		}
	}
	return errors
}

// AnalyzeFile evaluates one rule against one file,
// if the rule has children they will also be evaluated
func AnalyzeFile(rule *types.Rule, env *types.Env) (*types.Match, []error) {
	reader := env.Input
	reader.Seek(0, os.SEEK_SET)

	// 1. evaluate the rule
	match, err := exp.RuleEval(rule, env)
	if err != nil {
		// no need to record errors at this stage?
		return nil, nil
	}
	if !match {
		return nil, nil
	}

	// 2. perform actions
	errors := executeRuleActions(rule, env)

	// 3. record the match
	vars := exp.ScopeExtract(env.Scope)
	m := &types.Match{Rule: rule, Vars: vars}

	// 4. call children
	for _, cr := range rule.Children {
		env.PushRule(cr)
		cm, errs := AnalyzeFile(cr, env)
		if cm != nil {
			m.Children = append(m.Children, cm)
			cm.Parent = m
		} else {
			m.FailedChildren = append(m.FailedChildren, cr)
		}
		errors = append(errors, errs...) // record errors from this
		env.PopRule()
	}

	return m, errors
}
