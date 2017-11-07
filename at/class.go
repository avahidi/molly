package at

import "fmt"

type Assignment struct {
	Id   string
	Expr Expression
}

type Condition struct {
	Expr Expression
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
		eb, okay := e.(*BooleanExpression)
		if !okay {
			return false, fmt.Errorf("condition is not a logic expression")
		}
		if !eb.Value {
			return eb.Value, nil
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
