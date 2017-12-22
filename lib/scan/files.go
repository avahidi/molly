package scan

import (
	"fmt"

	"bitbucket.org/vahidi/molly/lib/types"
)

// AnalyzeFile evaluates one rule against one file,
// if the rule has children they will also be evaluated
func AnalyzeFile(filename string, r types.Rule, e types.Env) (*types.MatchEntry, []error) {
	// 1. evaluate the rule
	match, err := r.Eval(e)
	if err != nil {
		// no need to record errors at this stage?
		return nil, nil
	}
	if !match {
		return nil, nil
	}

	// 2. call all action functions
	var errors []error
	r.Actions(func(r1 types.Rule, a types.Expression) error {
		if _, err := a.Eval(e); err != nil {
			err := fmt.Errorf("[action failure] %.100q...", err)
			errors = append(errors, err)
		}
		return err
	})

	// 3. record the match
	m := &types.MatchEntry{Filename: filename, Rule: r.GetId(), Vars: e.Extract()}
	// s.Results = append(s.Results, m)

	// 4. call children
	r.Children(func(cr types.Rule) error {
		e.PushScope(cr)
		cm, errs := AnalyzeFile(filename, cr, e)
		if cm != nil {
			m.Children = append(m.Children, cm)
			cm.Parent = m
		}
		errors = append(errors, errs...) // record errors from this
		e.PopScope()
		return nil
	})

	return m, errors
}
