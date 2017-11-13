package at

import (
	"fmt"
	"os"

	"bitbucket.org/vahidi/molly/prim"
)

type Expression interface {
	Eval(env *Env) (Expression, error)
	Get() interface{}
}

// type assertions
var _ Expression = (*ValueExpression)(nil)
var _ Expression = (*VariableExpression)(nil)
var _ Expression = (*BinaryExpression)(nil)
var _ Expression = (*UnaryExpression)(nil)
var _ Expression = (*ExtractExpression)(nil)

type ValueExpression struct {
	Value prim.Primitive
}

func NewValueExpression(val prim.Primitive) *ValueExpression {
	return &ValueExpression{Value: val}
}

func NewNumberExpression(val uint64, size int, format prim.Format) *ValueExpression {
	nn := prim.NewNumber(size, format)
	nn.Set(val)
	return NewValueExpression(nn)
}

func NewBooleanExpression(val bool) *ValueExpression {
	n := uint64(0)
	if val {
		n = ^n
	}
	return NewNumberExpression(n, 8, prim.UBE)
}
func NewStringExpression(s string) *ValueExpression {
	ss := prim.NewString()
	ss.SetString(s)
	return NewValueExpression(ss)
}

func (ve *ValueExpression) Eval(env *Env) (Expression, error) {
	return ve, nil
}

func (ve ValueExpression) Get() interface{} {
	return ve.Value
}

func (ve ValueExpression) String() string {
	return fmt.Sprintf("%v", ve.Value)
}

// variable expression
type VariableExpression struct {
	Id string
}

func (ve *VariableExpression) Eval(env *Env) (Expression, error) {
	if env == nil {
		return ve, nil
	}
	expr := env.Scope.Get(ve.Id)
	if expr == nil {
		return nil, fmt.Errorf("Could not find variable %s", ve.Id)
	}
	return expr, nil // TODO
}

func (ve VariableExpression) Get() interface{} {
	return &ve
}

func (ve *VariableExpression) String() string {
	return ve.Id
}

// binary expression
type BinaryExpression struct {
	Left      Expression
	Right     Expression
	Operation prim.Operation
}

func (be *BinaryExpression) Eval(env *Env) (Expression, error) {
	left, err := be.Left.Eval(env)
	if err != nil {
		return nil, err
	}
	right, err := be.Right.Eval(env)
	if err != nil {
		return nil, err
	}

	// we need values to do more
	v1, okay1 := left.(*ValueExpression)
	v2, okay2 := right.(*ValueExpression)
	if !okay1 || !okay2 {
		if env == nil {
			return be, nil
		}
		return nil, fmt.Errorf("Expected values in binary operation")
	}

	k, err := v1.Value.Binary(v2.Value, be.Operation)
	return NewValueExpression(k), err
}

func (be BinaryExpression) Get() interface{} {
	return &be
}

// unary expression
type UnaryExpression struct {
	Value     Expression
	Operation prim.Operation
}

func (ue *UnaryExpression) Eval(env *Env) (Expression, error) {
	val, err := ue.Value.Eval(env)
	if err != nil {
		return nil, err
	}

	// we need values to do more
	v1, okay1 := val.(*ValueExpression)
	if !okay1 {
		if env == nil {
			return ue, nil
		}
		return nil, fmt.Errorf("Expected values in unary operation")
	}

	switch n := v1.Value.(type) {
	case *prim.Number:
		k, err := n.Unary(ue.Operation)
		return NewValueExpression(k), err
	}
	return nil, fmt.Errorf("Unknown unary expression")
}

func (ue UnaryExpression) Get() interface{} {
	return &ue
}

// extract expression
type ExtractExpression struct {
	Offset Expression
	Size   Expression
	Format prim.Format
}

func (ee *ExtractExpression) Eval(env *Env) (Expression, error) {
	if env == nil {
		return ee, nil
	}
	o1, err := ee.Offset.Eval(env)
	if err != nil {
		return nil, err
	}

	s1, err := ee.Size.Eval(env)
	if err != nil {
		return nil, err
	}
	o := o1.Get().(*prim.Number)
	s := s1.Get().(*prim.Number)
	if _, err := env.file.Seek(int64(o.Value), os.SEEK_SET); err != nil {
		return nil, err
	}

	data := make([]byte, s.Value)
	if _, err := env.file.Read(data); err != nil {
		return nil, err
	}

	n := prim.NewNumber(int(s.Value), ee.Format)
	err = n.Extract(data)
	return NewValueExpression(n), err

}

func (ue ExtractExpression) Get() interface{} {
	return &ue
}
