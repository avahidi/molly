package at

import "fmt"

type Scope struct {
	variables map[string]Expression
	parent    *Scope
}

func NewScope() *Scope {
	return &Scope{
		variables: make(map[string]Expression),
	}
}

func (s Scope) Get(id string) (Expression, bool) {
	e, found := s.variables[id]
	if !found {
		if s.parent != nil {
			return s.parent.Get(id)
		}
		return nil, false
	}
	return e, true
}

func (s *Scope) Set(id string, exp Expression) {
	s.variables[id] = exp
}

func (e Scope) Dump() {
	for k, v := range e.variables {
		fmt.Printf("EXPR %s = %s\n", k, v)
	}
}

// Extract values from all expressions that have any
func (e Scope) Extract() map[string]interface{} {
	ret := make(map[string]interface{})
	extract(&e, ret)
	return ret
}

func extract(e *Scope, ret map[string]interface{}) {
	for k, v := range e.variables {
		if _, found := ret[k]; !found {
			if vv, fin := v.(*ValueExpression); fin {
				ret[k] = vv.Value.Get()
			}
		}
	}
	if e.parent != nil {
		extract(e.parent, ret)
	}
}
