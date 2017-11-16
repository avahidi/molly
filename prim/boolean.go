package prim

import "fmt"

type Boolean struct {
	Value bool
}

func NewBoolean(val bool) *Boolean {
	return &Boolean{Value: val}
}

/*
func (b *Boolean) Set(val bool) {
	b.Value = val
}
*/
func (n *Boolean) Binary(o Primitive, op Operation) (Primitive, error) {
	m := o.(*Boolean)

	// comparision
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

	return nil, fmt.Errorf("Unknown boolean binary operation: %v", op)
}

func (n *Boolean) Unary(op Operation) (Primitive, error) {
	switch op {
	case INV, NEG:
		return NewBoolean(!n.Value), nil
	default:
		return nil, fmt.Errorf("Unknown boolean  operation: %v", op)
	}
}

func (n *Boolean) Get() interface{} {
	return n.Value
}

// type assertion NumBooleanber -> Primitive
var _ Primitive = (*Boolean)(nil)
