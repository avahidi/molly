package scan

import (
	"fmt"
	"os"

	"bitbucket.org/vahidi/molly/lib/exp"
	"bitbucket.org/vahidi/molly/lib/types"
)

// AnalyzeFile evaluates one rule against one file,
// if the rule has children they will also be evaluated
func AnalyzeFile(rule *types.Rule, env *types.Env) (*types.MatchEntry, []error) {
	reader := env.Reader
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

	// 2. call all action functions
	var errors []error
	for _, a := range rule.Actions {
		if _, err := a.Eval(env); err != nil {
			err := fmt.Errorf("[action failure] %.100q...", err)
			errors = append(errors, err)
		}
	}

	// 3. record the match
	filename, _ := env.Globals.GetString("$filename", "")
	vars := exp.ScopeExtract(env.Scope)
	m := &types.MatchEntry{Filename: filename, Rule: rule.ID, Vars: vars}
	// s.Results = append(s.Results, m)

	// 4. call children
	for _, cr := range rule.Children {
		env.PushRule(cr)
		cm, errs := AnalyzeFile(cr, env)
		if cm != nil {
			m.Children = append(m.Children, cm)
			cm.Parent = m
		}
		errors = append(errors, errs...) // record errors from this
		env.PopRule()
	}

	return m, errors
}
