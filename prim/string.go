package prim

import (
	"fmt"
)

type String struct {
	Value []byte
}

func NewString() *String {
	return &String{}
}

func (n String) Clone() *String {
	ret := &String{}
	*ret = n
	return ret
}

func (n *String) SetString(s string) {
	n.Value = []byte(s)
}

func (n *String) Set(data []byte) {
	n.Value = make([]byte, len(data))
	copy(n.Value, data)
}

func (n *String) Extract(data []byte) error {
	n.Set(data)
	return nil
}

func (n String) String() string {
	return string(n.Value)
}

func (n String) equals(m String) bool {
	if len(n.Value) != len(m.Value) {
		return false
	}
	for i, v1 := range n.Value {
		if v1 != m.Value[i] {
			return false
		}
	}
	return true
}

func (n *String) Binary(o Primitive, op Operation) (Primitive, error) {
	m := o.(*String)

	switch op {
	case ADD:
		ret := n.Clone()
		b := make([]byte, len(n.Value)+len(m.Value))
		copy(b[0:len(n.Value)], n.Value)
		copy(b[len(n.Value):], m.Value)
		ret.Set(b)
		return ret, nil

	case EQ:
		ret := NewBoolean()
		ret.SetBoolean(n.equals(*m))
		return ret, nil
	case NE:
		ret := NewBoolean()
		ret.SetBoolean(!n.equals(*m))
		return ret, nil

	}

	return nil, fmt.Errorf("Unknown string binary operation: %v", op)
}

func (n *String) Unary(op Operation) (Primitive, error) {
	return nil, fmt.Errorf("Unknown string unary operation: %v", op)
}

func (n *String) Get() interface{} {
	return string(n.Value)
}

// type assertion String -> Primitive
var _ Primitive = (*String)(nil)
