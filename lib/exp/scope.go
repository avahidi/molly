package exp

import (
	"fmt"

	"bitbucket.org/vahidi/molly/lib/types"
)

type scope struct {
	variables map[string]types.Expression
	parent    *scope
	rule      types.Rule
}

var _ types.Scope = (*scope)(nil)

func newScope(rule types.Rule, parent *scope) *scope {
	return &scope{
		variables: make(map[string]types.Expression),
		parent:    parent,
		rule:      rule,
	}
}

func (s scope) GetParent() types.Scope { return s.parent }
func (s scope) GetRule() types.Rule    { return s.rule }

func (s scope) Get(id string) (types.Expression, bool) {
	e, found := s.variables[id]
	if !found {
		if s.parent != nil {
			return s.parent.Get(id)
		}
		return nil, false
	}
	return e, true
}

func (s *scope) Set(id string, exp types.Expression) {
	s.variables[id] = exp
}
func (s *scope) Delete(id string) {
	delete(s.variables, id)
}

func (e scope) Dump() {
	for k, v := range e.variables {
		fmt.Printf("EXPR %s = %s\n", k, v)
	}
}

// Extract values from all expressions that have any
func (e scope) Extract() map[string]interface{} {
	ret := make(map[string]interface{})
	for k, v := range e.variables {
		if _, found := ret[k]; !found {
			if vv, fin := v.(*ValueExpression); fin {
				ret[k] = vv.Value.Get()
			}
		}
	}
	return ret
}
