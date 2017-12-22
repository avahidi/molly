package scan

import (
	"fmt"

	"bitbucket.org/vahidi/molly/lib/exp"
	"bitbucket.org/vahidi/molly/lib/types"
)

// AnalyzeFile evaluates one rule against one file,
// if the rule has children they will also be evaluated
func AnalyzeFile(filename string, r *types.Rule, e *types.Env) (*types.MatchEntry, []error) {
	// 1. evaluate the rule
	match, err := exp.RuleEval(r, e)
	if err != nil {
		// no need to record errors at this stage?
		return nil, nil
	}
	if !match {
		return nil, nil
	}

	// 2. call all action functions
	var errors []error
	for _, a := range r.Actions {
		if _, err := a.Eval(e); err != nil {
			err := fmt.Errorf("[action failure] %.100q...", err)
			errors = append(errors, err)
		}
	}

	// 3. record the match
	m := &types.MatchEntry{Filename: filename, Rule: r.ID,
		Vars: exp.ScopeExtract(e.Scope)}
	// s.Results = append(s.Results, m)

	// 4. call children
	for _, cr := range r.Children {
		e.PushScope(cr)
		cm, errs := AnalyzeFile(filename, cr, e)
		if cm != nil {
			m.Children = append(m.Children, cm)
			cm.Parent = m
		}
		errors = append(errors, errs...) // record errors from this
		e.PopScope()
	}

	return m, errors
}
