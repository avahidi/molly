package exp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"

	"bitbucket.org/vahidi/molly/lib/util/logging"

	"bitbucket.org/vahidi/molly/lib/exp/prim"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
)

type rule struct {
	id          string
	metadata    *util.Register
	assignments map[string]types.Expression
	conditions  []types.Expression
	actions     []types.Expression
	parent      *rule
	children    []*rule
	closed      bool
}

var _ types.Rule = (*rule)(nil)

var (
	ErrorRuleNotOpen        = errors.New("Rule should be open but is closed")
	ErrorRuleNotClose       = fmt.Errorf("Rule should be closed for this operation")
	ErrorEnviormentNotValid = fmt.Errorf("Rule operation requires a valid enviorment")
)

func NewRule(id string) *rule {
	return &rule{
		id:          id,
		metadata:    util.NewRegister(),
		assignments: make(map[string]types.Expression),
	}
}

func (c rule) GetId() string               { return c.id }
func (c rule) GetMetadata() *util.Register { return c.metadata }

func (c *rule) Actions(f func(types.Rule, types.Expression) error) error {
	for _, a := range c.actions {
		if err := f(c, a); err != nil {
			return err
		}
	}
	return nil
}

func (c rule) Children(f func(types.Rule) error) error {
	for _, ch := range c.children {
		if err := f(ch); err != nil {
			return err
		}
	}
	return nil
}

func (c *rule) SetParent(p0 types.Rule) {
	if c.closed {
		logging.Fatalf("Cannot modify closed rule '%s'", c.id)
	}
	p := p0.(*rule)
	p.children = append(p.children, c)
	c.parent = p
	c.metadata.SetParent(p.metadata)
}

func (c *rule) AddVar(id string, e types.Expression) error {
	if c.closed {
		return ErrorRuleNotOpen
	}

	if _, exists := c.assignments[id]; exists {
		return fmt.Errorf("variable '%s' is already defined in %s", id, c.id)
	}
	c.assignments[id] = e
	return nil
}

func (c *rule) GetVar(id string) (types.Expression, bool) {
	e, found := c.assignments[id]
	return e, found
}

func (c *rule) AddCondition(e types.Expression) error {
	if c.closed {
		return ErrorRuleNotOpen
	}
	e = Simplify(e)
	c.conditions = append(c.conditions, e)
	return nil
}

func (c *rule) AddAction(action types.Expression) error {
	if c.closed {
		return ErrorRuleNotOpen
	}
	action = Simplify(action)
	c.actions = append(c.actions, action)
	return nil
}

func (c rule) GetActions() []types.Expression {
	return c.actions
}

func (c *rule) Close() {
	if c.closed {
		return
	}

	for id, a := range c.assignments {
		c.assignments[id] = Simplify(a)
	}

	for i := range c.conditions {
		c.conditions[i] = Simplify(c.conditions[i])
	}

	for i := range c.actions {
		c.actions[i] = Simplify(c.actions[i])
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
			metadata.SetParent(c.metadata)
		}
		return adjustMetadataParent
	}
	c.visitExpressions(adjustMetadataParent)
	c.closed = true
}

func (c *rule) Eval(env types.Env) (bool, error) {
	if env == nil {
		return false, ErrorEnviormentNotValid
	}

	if !c.closed {
		return false, ErrorRuleNotOpen
	}

	for _, n := range c.conditions {
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
	for id, orgexp := range c.assignments {
		if _, found := env.Get(id); !found {
			if exp, err := orgexp.Eval(env); err == nil {
				env.Set(id, exp)
			} else {
				// TODO: how do we report an error that happens AFTER
				// conditions have been met?
				return true, err
			}
		}
	}

	return true, nil
}

func (c rule) String() string {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	fmt.Fprintf(w, "Rule %s {\n", c.id)
	for id, a := range c.assignments {
		fmt.Fprintf(w, "\tvar %s = %s;\n", id, a)
	}
	for _, c := range c.conditions {
		fmt.Fprintf(w, "\tif %s;\n", c)
	}

	for _, a := range c.actions {
		fmt.Fprintf(w, "\taction %s;\n", a)
	}
	fmt.Fprintf(w, "}\n")
	w.Flush()
	return buf.String()
}

func (c *rule) visitExpressions(v visitor) {
	for _, a := range c.assignments {
		walk(a, v)
	}
	for _, cc := range c.conditions {
		walk(cc, v)
	}
	for _, a := range c.actions {
		walk(a, v)
	}
}
