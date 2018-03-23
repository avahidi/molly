package exp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"bitbucket.org/vahidi/molly/lib/exp/prim"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
)

func get(e types.Expression) interface{} {
	switch n := e.(type) {
	case *ValueExpression:
		return n.Value
	default:
		return e
	}
}

func Simplify(e types.Expression) types.Expression {
	sim, err := e.Eval(nil)
	if err == nil && sim != nil {
		return sim
	}
	return e
}

func requireNumberPrimitive(e types.Expression) (int, error) {
	ve, valid := e.(*ValueExpression)
	if !valid {
		return 0, fmt.Errorf("'%s' is not a value types.Expression", e)
	}

	iv := ve.Value.Get()
	switch n := iv.(type) {
	case int:
		return n, nil
	case int64:
		return int(n), nil
	case uint64:
		return int(n), nil
	default:
		return 0, fmt.Errorf("'%v' is not a number (%T)", iv, iv)
	}
}

func requireStringPrimitive(e types.Expression) ([]byte, error) {
	ve, valid := e.(*ValueExpression)
	if !valid {
		return nil, fmt.Errorf("'%s' is not a value types.Expression", e)
	}

	switch n := ve.Value.(type) {
	case *prim.String:
		return n.Value, nil
	default:
		return nil, fmt.Errorf("'%v' is not a string (%t %T)", ve, ve, ve)
	}
}

// type assertions
var _ types.Expression = (*ValueExpression)(nil)
var _ types.Expression = (*VariableExpression)(nil)
var _ types.Expression = (*OperationExpression)(nil)
var _ types.Expression = (*SliceExpression)(nil)
var _ types.Expression = (*ExtractExpression)(nil)
var _ types.Expression = (*FunctionExpression)(nil)

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

func (ve *ValueExpression) Eval(env *types.Env) (types.Expression, error) {
	return ve, nil
}

func (ve ValueExpression) String() string {
	return fmt.Sprintf("%v", ve.Value)
}

// variable types.Expression
type VariableExpression struct {
	Id string
}

func (ve *VariableExpression) Eval(env *types.Env) (types.Expression, error) {
	if env == nil {
		return ve, nil
	}
	expr, found, err := EnvLookup(env, ve.Id)
	if err != nil {
		return nil, err
	}
	if !found {
		err := fmt.Errorf("Could not find variable '%s'", ve.Id)
		util.RegisterFatal(err) // no point continuing after this...
		return nil, err
	}
	return expr, nil
}

func (ve *VariableExpression) String() string {
	return ve.Id
}

// SliceExpression contains an index/slice expresion
type SliceExpression struct {
	Expr  types.Expression
	Start types.Expression
	End   types.Expression
}

func (se *SliceExpression) Eval(env *types.Env) (types.Expression, error) {
	expr, err := se.Expr.Eval(env)
	if err != nil {
		return nil, err
	}

	start, err := se.Start.Eval(env)
	if err != nil {
		return nil, err
	}

	var end types.Expression
	if se.End != nil {
		if end, err = se.End.Eval(env); err != nil {
			return nil, err
		}
	}

	if env == nil {
		return NewSliceExpression(expr, start, end), nil
	}
	x1, err := requireStringPrimitive(expr)
	if err != nil {
		return nil, err
	}

	// left:
	s1, err := requireNumberPrimitive(start)
	if err != nil {
		return nil, err
	}
	if s1 < 0 || s1 >= len(x1) {
		return nil, fmt.Errorf("Index start out of range: %d", s1)
	}

	if end == nil {
		prim := prim.ValueToPrimitive(x1[s1])
		return NewValueExpression(prim), nil
	}

	// right
	e1, err := requireNumberPrimitive(end)
	if err != nil {
		return nil, err
	}

	if e1 <= s1 || e1 > len(x1) {
		return nil, fmt.Errorf("Index end out of range: %d", e1)
	}

	prim := prim.NewStringRaw(x1[s1:e1])
	return NewValueExpression(prim), nil
}

func NewSliceExpression(expr, start, end types.Expression) *SliceExpression {
	return &SliceExpression{Expr: expr, Start: start, End: end}
}

func (se SliceExpression) String() string {
	if se.End == nil {
		return fmt.Sprintf("%s[%s]", se.Expr, se.Start)
	}
	return fmt.Sprintf("%s[%s:%s]", se.Expr, se.Start, se.End)
}

// operation types.Expression
type OperationExpression struct {
	Left      types.Expression
	Right     types.Expression
	Operation prim.Operation
}

func (oe *OperationExpression) Eval(env *types.Env) (types.Expression, error) {
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
		return nil, fmt.Errorf("Unknown unary types.Expression")
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

func NewBinaryExpression(left, right types.Expression,
	op prim.Operation) *OperationExpression {
	return &OperationExpression{Left: left, Right: right, Operation: op}
}

func NewUnaryExpression(expr types.Expression, op prim.Operation) *OperationExpression {
	return &OperationExpression{Left: expr, Right: nil, Operation: op}
}

func (oe OperationExpression) String() string {
	if oe.Right == nil {
		return fmt.Sprintf("(%s %s)", oe.Operation, oe.Left)
	}
	return fmt.Sprintf("(%s %s %s)", oe.Left, oe.Operation, oe.Right)
}

// extract types.Expression

const (
	UNumber = iota
	SNumber
	String
	StringZ
)

const (
	EndianUnknown = iota
	EndianBig
	EndianLittle
)

type ExtractFormat struct {
	Type  int
	Order int
}

func (e ExtractFormat) Signed() bool {
	return e.Type == SNumber
}

func (e *ExtractFormat) SetByteOrder(bigendian bool) {
	if bigendian {
		e.Order = EndianBig
	} else {
		e.Order = EndianLittle
	}
}
func (e ExtractFormat) ByteOrder() binary.ByteOrder {
	if e.Order == EndianLittle {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

type ExtractExpression struct {
	Offset   types.Expression
	Size     types.Expression
	Format   ExtractFormat
	Metadata *util.Register
}

func NewExtractExpression(o, s types.Expression, m *util.Register, f ExtractFormat) *ExtractExpression {
	return &ExtractExpression{Offset: o, Size: s, Metadata: m, Format: f}
}

func (ee *ExtractExpression) Eval(env *types.Env) (types.Expression, error) {
	o1, err := ee.Offset.Eval(env)
	if err != nil {
		return nil, err
	}
	s1, err := ee.Size.Eval(env)
	if err != nil {
		return nil, err
	}

	// simplified?
	if env == nil {
		return NewExtractExpression(o1, s1, ee.Metadata, ee.Format), nil
	}

	o := get(o1).(*prim.Number)
	s := get(s1).(*prim.Number)
	if _, err := env.Input.Seek(int64(o.Value), os.SEEK_SET); err != nil {
		return nil, err
	}

	// endian is not known yet, see if we can figure it out from metadata
	if ee.Format.Order == EndianUnknown {
		be, _ := ee.Metadata.GetBoolean("bigendian", true)
		ee.Format.SetByteOrder(be)
	}

	if ee.Format.Type == String || ee.Format.Type == StringZ {
		var data []byte
		var err error

		// zero terminated or fix size?
		if ee.Format.Type == StringZ {
			data, _, err = util.ReadUntil(env.Input, 0, int(s.Value))
		} else {
			var n int
			data = make([]byte, s.Value)
			n, err = env.Input.Read(data)
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
		if _, err := env.Input.Read(data); err != nil {
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
			val = uint64(bo.Uint64(data))
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

type FunctionExpression struct {
	Name     string
	Func     *types.Function
	Params   []types.Expression
	Metadata *util.Register
}

func NewFunctionExpression(name string, m *util.Register,
	params ...types.Expression) (*FunctionExpression, error) {
	f, found := types.FunctionFind(name)
	if !found {
		return nil, fmt.Errorf("Unknown function: %s", name)
	}
	return &FunctionExpression{Name: name, Metadata: m, Func: f, Params: params}, nil
}

func (fe *FunctionExpression) Eval(env *types.Env) (types.Expression, error) {
	// simplify:
	if env == nil {
		c := &FunctionExpression{Name: fe.Name, Func: fe.Func,
			Params: make([]types.Expression, len(fe.Params))}
		for i, p1 := range fe.Params {
			if p2, err := p1.Eval(nil); err == nil {
				c.Params[i] = p2
			}
		}
		return c, nil
	}

	// get values instead of types.Expressions
	ps := make([]interface{}, len(fe.Params))
	for i, p1 := range fe.Params {
		p2, err := p1.Eval(env)
		if err != nil {
			return nil, err
		}
		v2, okay := p2.(*ValueExpression)
		if !okay {
			return nil, fmt.Errorf("function parameter is not a primitive: %v", p2)
		}
		ps[i] = v2.Value.Get()
	}

	r, err := fe.Func.Call(env, ps)
	if err != nil {
		return nil, err
	}
	return NewValueExpression(prim.ValueToPrimitive(r)), nil

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
