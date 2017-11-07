package at

import (
	"fmt"
	"os"
)

type Expression interface {
	Eval(env *Env) (Expression, error)
}

type BooleanExpression struct {
	Value bool
}

func NewBooleanExpression(value bool) *BooleanExpression {
	return &BooleanExpression{Value: value}
}

func (be *BooleanExpression) Eval(env *Env) (Expression, error) {
	return be, nil
}

func (be *BooleanExpression) String() string {
	if be.Value {
		return "true"
	}
	return "false"
}

// variable expression
type VariableExpression struct {
	Id string
}

func (ve *VariableExpression) Eval(env *Env) (Expression, error) {
	expr := env.Scope.Get(ve.Id)
	if expr == nil {
		return nil, fmt.Errorf("Could not find variable %s", ve.Id)
	}
	return expr, nil // TODO
}

func (ve *VariableExpression) String() string {
	return ve.Id
}

// number expression
type NumberExpression struct {
	Value uint64
}

func (ne *NumberExpression) Eval(env *Env) (Expression, error) {
	return ne, nil
}

func (ne *NumberExpression) String() string {
	return fmt.Sprintf("0x%x", ne.Value)
}

// binary expression
type BinaryExpression struct {
	Left      Expression
	Right     Expression
	Operation Operation
}

func binaryNumberOperation(l *NumberExpression, r *NumberExpression,
	op Operation) (Expression, error) {
	switch op {
	case ADD:
		return &NumberExpression{Value: l.Value + r.Value}, nil
	case SUB:
		return &NumberExpression{Value: l.Value - r.Value}, nil
	case DIV:
		if r.Value == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return &NumberExpression{Value: l.Value / r.Value}, nil
	case MUL:
		return &NumberExpression{Value: l.Value * r.Value}, nil

	case AND:
		return &NumberExpression{Value: l.Value & r.Value}, nil
	case OR:
		return &NumberExpression{Value: l.Value | r.Value}, nil
	case XOR:
		return &NumberExpression{Value: l.Value ^ r.Value}, nil

	case EQ:
		return &BooleanExpression{Value: l.Value == r.Value}, nil
	case NE:
		return &BooleanExpression{Value: l.Value != r.Value}, nil
	case GT:
		return &BooleanExpression{Value: l.Value > r.Value}, nil
	case LT:
		return &BooleanExpression{Value: l.Value < r.Value}, nil

		/*
			case BAND:
				return &BooleanExpression{Value: l.Value > r.Value}, nil
			case BOR:
				return &BooleanExpression{Value: l.Value < r.Value}, nil
			case BXOR:
				return &BooleanExpression{Value: l.Value > r.Value}, nil
		*/
	}
	return nil, fmt.Errorf("Unknown number operation: %v", op)
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

	switch n := left.(type) {
	case *NumberExpression:
		switch m := right.(type) {
		case *NumberExpression:
			return binaryNumberOperation(n, m, be.Operation)
		}

	}
	return nil, fmt.Errorf("Unknown binary expression")
}

// extract expression
type ExtractExpression struct {
	Offset Expression
	Size   Expression
	Format Format
}

func (ee *ExtractExpression) Eval(env *Env) (Expression, error) {
	o1, err := ee.Offset.Eval(env)
	if err != nil {
		return nil, err
	}

	s1, err := ee.Size.Eval(env)
	if err != nil {
		return nil, err
	}
	o := o1.(*NumberExpression)
	s := s1.(*NumberExpression)
	if _, err := env.file.Seek(int64(o.Value), os.SEEK_SET); err != nil {
		return nil, err
	}

	data := make([]byte, s.Value)
	if _, err := env.file.Read(data); err != nil {
		return nil, err
	}

	// Assuming its (little endian, unsigned, number ) expression for now
	ret := &NumberExpression{}
	// TODO ret.Format = ee.Format
	ret.Value = ee.Format.Extract(len(data), data)

	return ret, nil
}
