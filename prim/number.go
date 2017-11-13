package prim

import (
	"encoding/binary"
	"fmt"
)

type Format int

const (
	UBE Format = iota
	SBE
	ULE
	SLE
)

var formatNames = [...]string{
	UBE: "UBE",
	SBE: "SBE",
	ULE: "ULE",
	SLE: "SLE",
}

func (f Format) ByteOrder() binary.ByteOrder {
	if f == UBE || f == SBE {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

func (f Format) Signed() bool {
	return f == SBE || f == SLE
}

func (f Format) String() string {
	return formatNames[int(f)]
}

type Number struct {
	Value  uint64
	Format Format
	Size   int
}

func NewNumber(size int, format Format) *Number {
	return &Number{Format: format, Size: size}
}

func NewBoolean() *Number {
	return NewNumber(4, UBE)
}

func (n Number) Clone() *Number {
	ret := &Number{}
	*ret = n
	return ret
}

func (n *Number) Set(val uint64) {
	if n.Format.Signed() {
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
}

func (n *Number) SetBoolean(val bool) {
	n.Value = 0
	if val {
		n.Value = ^n.Value
	}
}

func (n *Number) Extract(data []byte) error {
	bo := n.Format.ByteOrder()
	val := uint64(0)
	switch n.Size {
	case 1:
		val = uint64(data[0])
	case 2:
		val = uint64(bo.Uint16(data))
	case 4:
		val = uint64(bo.Uint32(data))
	case 8:
		val = bo.Uint64(data)
	default:
		return fmt.Errorf("Internal error: invalid number length: %d", n.Size)
	}

	n.Set(val)
	return nil
}

func (n *Number) Binary(o Primitive, op Operation) (Primitive, error) {
	m := o.(*Number)
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
	case MUL:
		ret.Set(n.Value * m.Value)
	case AND:
		ret.Set(n.Value & m.Value)
	case OR:
		ret.Set(n.Value | m.Value)
	case XOR:
		ret.Set(n.Value ^ m.Value)

	case EQ:
		ret = NewBoolean()
		ret.SetBoolean(n.Value == m.Value)
	case NE:
		ret = NewBoolean()
		ret.SetBoolean(n.Value != m.Value)
	case GT:
		ret = NewBoolean()
		ret.SetBoolean(n.Value > m.Value)
	case LT:
		ret = NewBoolean()
		ret.SetBoolean(n.Value < m.Value)
	default:
		return nil, fmt.Errorf("Unknown number binary operation: %v", op)
	}
	return ret, nil
}

func (n *Number) Unary(op Operation) (Primitive, error) {
	ret := n.Clone()
	switch op {
	case NEG:
		ret.Set(-n.Value)
	case INV:
		ret.Set(^n.Value)
	default:
		return nil, fmt.Errorf("Unknown number unary operation: %v", op)
	}
	return ret, nil
}

func (n *Number) Get() interface{} {

	switch n.Format {
	case UBE, ULE:
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
	default:
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
