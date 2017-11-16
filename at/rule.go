package at

import (
	"bufio"
	"bytes"
	"fmt"

	"bitbucket.org/vahidi/molly/prim"
)

type Assignment struct {
	Id   string
	Expr Expression
}

type Condition struct {
	Expr Expression
}

type Action struct {
	Action Expression
}

type Rule struct {
	Id          string
	Assignments []*Assignment
	Conditions  []*Condition
	Actions     []*Action

	Parent   *Rule
	Children []*Rule
}

func NewRule(id string) *Rule {
	return &Rule{Id: id}
}

func (c *Rule) AddAssignment(id string, e Expression) {
	c.Assignments = append(c.Assignments, &Assignment{Id: id, Expr: e})
}

func (c *Rule) AddCondition(e Expression) {
	c.Conditions = append(c.Conditions, &Condition{Expr: e})
}

func (c *Rule) AddAction(action Expression) {
	a := &Action{Action: action}
	c.Actions = append(c.Actions, a)
}

func (c *Rule) Simplify() {
	for _, a := range c.Assignments {
		sim, err := a.Expr.Eval(nil)
		if err == nil && sim != nil {
			a.Expr = sim
		}
	}

	for _, c := range c.Conditions {
		sim, err := c.Expr.Eval(nil)
		if err == nil && sim != nil {
			c.Expr = sim
		}
	}

	for _, a := range c.Actions {
		sim, err := a.Action.Eval(nil)
		if err == nil && sim != nil {
			a.Action = sim
		}
	}

	for _, ch := range c.Children {
		ch.Simplify()
	}
}

func (c *Rule) Eval(env *Env) (bool, error) {
	for _, a := range c.Assignments {
		e, err := a.Expr.Eval(env)
		if err != nil {
			return false, err
		}
		env.Set(a.Id, e)
	}
	for _, n := range c.Conditions {
		e, err := n.Expr.Eval(env)
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

	return true, nil
}

func (c Rule) String() string {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	fmt.Fprintf(w, "Rule %s {\n", c.Id)
	for _, a := range c.Assignments {
		fmt.Fprintf(w, "\tvar %s = %s;\n", a.Id, a.Expr)
	}
	for _, c := range c.Conditions {
		fmt.Fprintf(w, "\tif %s;\n", c.Expr)
	}

	for _, a := range c.Actions {
		fmt.Fprintf(w, "\taction %s;\n", a.Action)
	}
	fmt.Fprintf(w, "}\n")
	w.Flush()
	return buf.String()
}
