package types

import (
	"bitbucket.org/vahidi/molly/lib/util"
)

// Expression is a node in the AST
type Expression interface {
	Eval(env *Env) (Expression, error)
}

// Scope is the current scope for a rule
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

// Rule defines a single rule
type Rule struct {
	ID       string
	Metadata *util.Register
	Parent   *Rule

	Children   []*Rule
	Conditions []Expression
	Actions    []Expression
	Variables  map[string]Expression
}

func NewRule(id string) *Rule {
	return &Rule{
		ID:        id,
		Metadata:  util.NewRegister(),
		Variables: make(map[string]Expression),
	}
}

// RuleSet represents a group of rules parsed from one or more file
// it also includes the rule hierarchy
type RuleSet struct {
	Files map[string][]*Rule
	Top   map[string]*Rule
	Flat  map[string]*Rule
}

func NewRuleSet() *RuleSet {
	return &RuleSet{
		Files: make(map[string][]*Rule),
		Top:   make(map[string]*Rule),
		Flat:  make(map[string]*Rule),
	}
}

// some helper functions to simplify code elsewhere

// FileName returns the name of the currently scanned binary file
func FileName(e *Env) string {
	filename, found := e.Globals.GetString("$filename", "")
	if !found {
		util.RegisterFatalf("internal error: could not find $filename (%s)", e)
	}
	return filename
}

// FileSize returns the size of the currently scanned binary file
func FileSize(e *Env) uint64 {
	size, found := e.Globals.GetNumber("$filesize", 0)
	if !found {
		util.RegisterFatalf("internal error: could not find $filesize (%s)", e)
	}
	return size
}
