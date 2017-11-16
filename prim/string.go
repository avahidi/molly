package prim

import (
	"bufio"
	"bytes"
	"fmt"
)

type String struct {
	Value []byte
}

func NewStringRaw(val []byte) *String {
	s := &String{}
	s.Set(val)
	return s
}

func NewString(str string) *String {
	return NewStringRaw([]byte(str))
}

func (n String) Clone() *String {
	ret := &String{}
	ret.Set(n.Value)
	return ret
}

func (n *String) SetString(s string) {
	n.Value = []byte(s)
}

func (n *String) Set(data []byte) {
	n.Value = make([]byte, len(data))
	copy(n.Value, data)
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

func (n String) isPrint() bool {
	for _, c := range n.Value {
		if c < ' ' || c >= 128 {
			return false
		}
	}
	return true
}
func (n String) String() string {
	if n.isPrint() {
		return string(n.Value)
	} else {
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)

		fmt.Fprintf(w, "{")
		for i, c := range n.Value {
			if i != 0 {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "0x%02x", c)
		}
		fmt.Fprintf(w, "}")
		w.Flush()
		return buf.String()
	}
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
		ret := NewBoolean(n.equals(*m))
		return ret, nil
	case NE:
		ret := NewBoolean(!n.equals(*m))
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
