package prim

import (
	"fmt"
)

type Number struct {
	Value  uint64
	Size   int
	Signed bool
}

func NewNumber(val uint64, size int, signed bool) *Number {
	n := &Number{Value: 0, Size: size, Signed: signed}
	n.Set(val)
	return n
}

func (n Number) Clone() *Number {
	ret := &Number{}
	*ret = n
	return ret
}

func (n *Number) Set(val uint64) *Number {
	if n.Signed {
		// sign extend
		mask := uint64(0xFFFFFFFFFFFFFF80) << uint64(8*(n.Size-1))
		if (mask & val) != 0 {
			val |= mask
		}
	} else {
		// remove unused bits
		mask := uint64(0xFFFFFFFFFFFFFFFF) << uint64(8*n.Size)
		val = val & ^mask
	}
	n.Value = val
	return n
}

func (n *Number) Binary(o Primitive, op Operation) (Primitive, error) {
	m := o.(*Number)

	// comparison
	switch op {
	case EQ:
		return NewBoolean(n.Value == m.Value), nil
	case NE:
		return NewBoolean(n.Value != m.Value), nil
	case GT:
		return NewBoolean(n.Value > m.Value), nil
	case GE:
		return NewBoolean(n.Value >= m.Value), nil
	case LE:
		return NewBoolean(n.Value <= m.Value), nil
	case LT:
		return NewBoolean(n.Value < m.Value), nil
	case BAND:
		return NewBoolean(n.Value != 0 && m.Value != 0), nil
	case BOR:
		return NewBoolean(n.Value != 0 || m.Value != 0), nil
	}

	// arith and logic
	ret := n.Clone()
	switch op {
	case ADD:
		ret.Set(n.Value + m.Value)
	case SUB:
		ret.Set(n.Value - m.Value)
	case DIV:
		if m.Value == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		ret.Set(n.Value / m.Value)
	case MOD:
		if m.Value == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		ret.Set(n.Value % m.Value)
	case MUL:
		ret.Set(n.Value * m.Value)
	case AND:
		ret.Set(n.Value & m.Value)
	case OR:
		ret.Set(n.Value | m.Value)
	case XOR:
		ret.Set(n.Value ^ m.Value)
	case LSL:
		ret.Set(n.Value << m.Value)
	case LSR:
		ret.Set(n.Value >> m.Value)
	default:
		return nil, fmt.Errorf("Unknown number binary operation: %v", op)
	}

	return ret, nil
}

func (n *Number) Unary(op Operation) (Primitive, error) {
	ret := n.Clone()
	switch op {
	case NEG, SUB:
		ret.Set(-n.Value)
	case INV:
		ret.Set(^n.Value)
	default:
		return nil, fmt.Errorf("Unknown number unary operation: %v", op)
	}
	return ret, nil
}

func (n *Number) Get() interface{} {

	if n.Signed {
		switch n.Size {
		case 1:
			return uint8(n.Value)
		case 2:
			return uint16(n.Value)
		case 4:
			return uint32(n.Value)
		default:
			return uint64(n.Value)
		}
	} else {
		switch n.Size {
		case 1:
			return int8(n.Value)
		case 2:
			return int16(n.Value)
		case 4:
			return int32(n.Value)
		default:
			return int64(n.Value)
		}
	}
}

// type assertion Number -> Primitive
var _ Primitive = (*Number)(nil)
