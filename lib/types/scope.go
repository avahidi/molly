package types

// Scope is the current scope while scanning a file for some rule.
// Since rules are in hierarchy, so does the scope
type Scope struct {
	Rule      *Rule
	Parent    *Scope
	variables map[string]Expression
}

// Get reads a variable from scope or parent scope
func (s Scope) Get(id string) (Expression, bool) {
	expr, found := s.variables[id]
	if !found && s.Parent != nil {
		expr, found = s.Parent.Get(id)
	}
	return expr, found
}

// Set writes a variable to the scope
func (s *Scope) Set(id string, e Expression) {
	s.variables[id] = e
}

// GetAll returns all scope variables
func (s Scope) GetAll() map[string]Expression { return s.variables }

// NewScope creates a new scope for a rule
func NewScope(rule *Rule, parent *Scope) *Scope {
	return &Scope{
		variables: make(map[string]Expression),
		Parent:    parent,
		Rule:      rule,
	}
}
