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

func (c *Condition) Simplify() {
	e2, err := c.Expr.Eval(nil)
	if err != nil && e2 != nil {
		c.Expr = e2
	}
}

type Action struct {
	Func ActionFunction
	Args []string
}

type Class struct {
	Id          string
	Assignments []*Assignment
	Conditions  []*Condition
	Actions     []*Action
}

func NewClass(id string) *Class {
	return &Class{Id: id}
}

func (c *Class) AddAssignment(id string, e Expression) {
	c.Assignments = append(c.Assignments, &Assignment{Id: id, Expr: e})
}

func (c *Class) AddCondition(e Expression) {
	c.Conditions = append(c.Conditions, &Condition{Expr: e})
}

func (c *Class) AddAction(id string, args ...string) {
	f := ActionRegister[id]
	c.Actions = append(c.Actions, &Action{Func: f, Args: args})
}

func (c *Class) Simplify() {
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
}

func (c *Class) Eval(env *Env) (bool, error) {
	for _, a := range c.Assignments {
		e, err := a.Expr.Eval(env)
		if err != nil {
			return false, err
		}
		env.Scope.Set(a.Id, e)
	}
	for _, n := range c.Conditions {
		e, err := n.Expr.Eval(env)
		if err != nil {
			return false, err
		}

		ve, okay := e.(*ValueExpression)
		if !okay {
			return false, fmt.Errorf("condition is not a value expression")
		}
		ne, okay1 := ve.Value.(*prim.Number)
		if !okay1 {
			return false, fmt.Errorf("condition is not a number expression")
		}
		if ne.Value == 0 {
			return false, nil
		}
	}

	for _, a := range c.Actions {
		err := a.Func(env, a.Args...)
		if err != nil {
			return true, err
		}
	}
	return true, nil
}

func (c Class) String() string {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	fmt.Fprintf(w, "Class %s {\n", c.Id)
	for _, a := range c.Assignments {
		fmt.Fprintf(w, "\tvar %s = %s;\n", a.Id, a.Expr)
	}
	for _, c := range c.Conditions {
		fmt.Fprintf(w, "\tif %s;\n", c.Expr)
	}

	for _, a := range c.Actions {
		fmt.Fprintf(w, "\taction %s %v;\n", a.Func, a.Args)
	}
	fmt.Fprintf(w, "}\n")
	w.Flush()
	return buf.String()
}
