package at

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"

	"bitbucket.org/vahidi/molly/prim"
	"bitbucket.org/vahidi/molly/util"
)

type Expression interface {
	Eval(env *Env) (Expression, error)
	Get() interface{}
}

// type assertions
var _ Expression = (*ValueExpression)(nil)
var _ Expression = (*VariableExpression)(nil)
var _ Expression = (*OperationExpression)(nil)
var _ Expression = (*ExtractExpression)(nil)
var _ Expression = (*FunctionExpression)(nil)

type ValueExpression struct {
	Value prim.Primitive
}

func NewValueExpression(val prim.Primitive) *ValueExpression {
	return &ValueExpression{Value: val}
}

func NewNumberExpression(val uint64, size int, signed bool) *ValueExpression {
	nn := prim.NewNumber(val, size, signed)
	return NewValueExpression(nn)
}

func NewBooleanExpression(val bool) *ValueExpression {
	nn := prim.NewBoolean(val)
	return NewValueExpression(nn)
}

func NewStringExpressionRaw(data []byte) *ValueExpression {
	ss := prim.NewStringRaw(data)
	return NewValueExpression(ss)
}

func NewStringExpression(s string) *ValueExpression {
	ss := prim.NewString(s)
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
	expr, found := env.Get(ve.Id)
	if !found {
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

// operation expression
type OperationExpression struct {
	Left      Expression
	Right     Expression
	Operation prim.Operation
}

func (oe *OperationExpression) Eval(env *Env) (Expression, error) {
	left, err := oe.Left.Eval(env)
	if err != nil {
		return nil, err
	}
	v1, okay1 := left.(*ValueExpression)
	if !okay1 {
		if env == nil {
			return oe, nil
		}
		return nil, fmt.Errorf("Invalid LHS in operation")
	}

	if oe.Right == nil {
		// UNARY
		switch n := v1.Value.(type) {
		case *prim.Number:
			k, err := n.Unary(oe.Operation)
			return NewValueExpression(k), err
		case *prim.Boolean:
			k, err := n.Unary(oe.Operation)
			return NewValueExpression(k), err
		}
		return nil, fmt.Errorf("Unknown unary expression")
	} else {
		// binary
		right, err := oe.Right.Eval(env)
		if err != nil {
			return nil, err
		}
		v2, okay2 := right.(*ValueExpression)
		if !okay2 {
			if env == nil {
				return oe, nil
			}
			return nil, fmt.Errorf("Invalid RHS in operation")
		}

		k, err := v1.Value.Binary(v2.Value, oe.Operation)
		return NewValueExpression(k), err
	}
}

func (oe OperationExpression) Get() interface{} {
	return &oe
}

func NewBinaryExpression(left, right Expression,
	op prim.Operation) *OperationExpression {
	return &OperationExpression{Left: left, Right: right, Operation: op}
}

func NewUnaryExpression(expr Expression, op prim.Operation) *OperationExpression {
	return &OperationExpression{Left: expr, Right: nil, Operation: op}
}

func (oe OperationExpression) String() string {
	if oe.Right == nil {
		return fmt.Sprintf("(%s %s)", oe.Operation, oe.Left)
	}
	return fmt.Sprintf("(%s %s %s)", oe.Left, oe.Operation, oe.Right)
}

// extract expression
type ExtractFormat int

const (
	NumberBEU ExtractFormat = iota
	NumberLEU
	NumberBES
	NumberLES
	String
	StringZ
)

func (e ExtractFormat) Signed() bool {
	return e == NumberBES || e == NumberLES
}

func (e ExtractFormat) ByteOrder() binary.ByteOrder {
	if e == NumberBES || e == NumberBEU {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

type ExtractExpression struct {
	Offset Expression
	Size   Expression
	Format ExtractFormat
}

func NewExtractExpression(o, s Expression, f ExtractFormat) *ExtractExpression {
	return &ExtractExpression{Offset: o, Size: s, Format: f}
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
	if err := env.Seek(int64(o.Value)); err != nil {
		return nil, err
	}

	if ee.Format == String || ee.Format == StringZ {
		var data []byte
		var err error

		// zero terminated or fix size?
		if ee.Format == StringZ {
			data, err = util.ReadUntil(env, 0, int(s.Value))
		} else {
			var n int
			data = make([]byte, s.Value)
			n, err = env.Read(data)
			if n != len(data) {
				err = fmt.Errorf("premature end of file in string")
			}
		}

		if err != nil {
			return nil, err
		}
		s := prim.NewStringRaw(data)
		return NewValueExpression(s), nil

	} else {
		data := make([]byte, s.Value)
		if _, err := env.Read(data); err != nil {
			return nil, err
		}
		val := uint64(0)
		bo := ee.Format.ByteOrder()

		switch len(data) {
		case 1:
			val = uint64(data[0])
		case 2:
			val = uint64(bo.Uint16(data))
		case 4:
			val = uint64(bo.Uint32(data))
		case 8:
			val = bo.Uint64(data)
		default:
			return nil, fmt.Errorf("Internal error: invalid number length: %d", len(data))
		}

		n := prim.NewNumber(val, len(data), ee.Format.Signed())
		return NewValueExpression(n), nil
	}
}

func (ue ExtractExpression) String() string {
	return fmt.Sprintf("Extract(%s, %s, %d)", ue.Offset, ue.Size, ue.Format)
}

func (ue ExtractExpression) Get() interface{} {
	return &ue
}

type FunctionExpression struct {
	Name   string
	Func   *ActionFunction
	Params []Expression
}

func NewFunctionExpression(name string, params ...Expression) (*FunctionExpression, error) {
	f, found := FindActionFunction(name)
	if !found {
		return nil, fmt.Errorf("Unknown function: %s", name)
	}
	return &FunctionExpression{Name: name, Func: f, Params: params}, nil
}

func (fe *FunctionExpression) Eval(env *Env) (Expression, error) {
	// simplify:
	if env == nil {
		c := &FunctionExpression{Name: fe.Name, Func: fe.Func,
			Params: make([]Expression, len(fe.Params))}
		for i, p1 := range fe.Params {
			if p2, err := p1.Eval(nil); err == nil {
				c.Params[i] = p2
			}
		}
		return c, nil
	}

	// get values instead of expressions
	ps := make([]interface{}, len(fe.Params))
	for i, p1 := range fe.Params {
		if p2, err := p1.Eval(env); err == nil {
			v2, okay := p2.(*ValueExpression)
			if !okay {
				return nil, fmt.Errorf("function parameter is not a primitive: %v", p2)
			}
			ps[i] = v2.Value.Get()
		}
	}

	r, err, efatal := fe.Func.Call(env, ps)
	if efatal != nil {
		panic(fmt.Sprintf("Fatal error when calling %s: %s", fe.Func, efatal))
	}
	if err != nil {
		return nil, err
	}
	return NewValueExpression(prim.ValueToPrimitive(r)), nil

}

func (fe FunctionExpression) Get() interface{} {
	return &fe
}

func (fe FunctionExpression) String() string {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	fmt.Fprintf(w, "%s ( ", fe.Name)
	for i, p := range fe.Params {
		if i != 0 {
			fmt.Fprintf(w, ", ")
		}
		fmt.Fprintf(w, "%s", p)
	}
	fmt.Fprintf(w, ")")
	w.Flush()
	return buf.String()
}
