package types

import "bitbucket.org/vahidi/molly/lib/util"

const (
	ActionModeNormal = 0
	ActionModeIgnore = 1
	ActionModeExit   = 2
)

type Action struct {
	Mode   int
	Action Expression
}

// Rule defines a single rule
type Rule struct {
	ID       string
	Metadata *util.Register
	Parent   *Rule `json:"-"` // this will avoid circular marshalling

	Children   []*Rule
	Conditions []Expression
	Actions    []Action
	Variables  map[string]Expression
}

// NewRule creates a new rule with the given ID
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

// NewRuleSet creates a new set of rules, to be populated by a rule scanner
func NewRuleSet() *RuleSet {
	return &RuleSet{
		Files: make(map[string][]*Rule),
		Top:   make(map[string]*Rule),
		Flat:  make(map[string]*Rule),
	}
}
