package prim

import "fmt"

type Boolean struct {
	Value bool
}

func NewBoolean(val bool) *Boolean {
	return &Boolean{Value: val}
}

func (n *Boolean) Binary(o Primitive, op Operation) (Primitive, error) {
	// early termination - when the first operation is enough to get result
	switch op {
	case BAND, AND:
		if !n.Value {
			return NewBoolean(false), nil
		}
	case BOR, OR:
		if n.Value {
			return NewBoolean(true), nil
		}
	}

	m, isbool := o.(*Boolean)
	if isbool {
		switch op {
		case EQ:
			return NewBoolean(n.Value == m.Value), nil
		case NE, XOR, BXOR:
			return NewBoolean(n.Value != m.Value), nil

		case BAND, AND:
			return NewBoolean(n.Value && m.Value), nil
		case BOR, OR:
			return NewBoolean(n.Value || m.Value), nil
		}
	}

	return nil, fmt.Errorf("Unknown boolean binary operation: %v %v %v", n, op, o)
}

func (n *Boolean) Unary(op Operation) (Primitive, error) {
	switch op {
	case INV, NEG:
		return NewBoolean(!n.Value), nil
	default:
		return nil, fmt.Errorf("Unknown boolean  operation: %v %v", op, n)
	}
}

func (n *Boolean) Get() interface{} {
	return n.Value
}

// type assertion NumBooleanber -> Primitive
var _ Primitive = (*Boolean)(nil)
