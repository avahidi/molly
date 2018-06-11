package prim

import (
	"reflect"
	"testing"
)

func TestValueToPrimitive(t *testing.T) {
	var testdata = []struct {
		input  interface{}
		output Primitive
	}{
		{true, &Boolean{true}},
		{false, &Boolean{false}},
		{int(10), &Number{Value: 10, Size: 8, Signed: true}},
		{uint(11), &Number{Value: 11, Size: 8, Signed: false}},
		{int32(12), &Number{Value: 12, Size: 4, Signed: true}},
		{[]byte{'a', 'b', 'c'}, &String{Value: []byte{'a', 'b', 'c'}}},
		{"XYZ", &String{Value: []byte{'X', 'Y', 'Z'}}},
	}

	for _, test := range testdata {
		p := ValueToPrimitive(test.input)
		if !reflect.DeepEqual(test.output, p) {
			t.Errorf("Expected %v got %v", test.output, p)
		}
	}
}

func TestBinary(t *testing.T) {
	var testdata = []struct {
		a, b, c interface{}
		op      Operation
		err     bool
	}{
		{true, true, true, EQ, false},
		{true, false, true, OR, false},
		{true, false, true, BOR, false},
		{true, false, false, BAND, false},
		{"aa", "bb", "aabb", ADD, false},
		{10, 20, 30, ADD, false},
	}

	for _, test := range testdata {
		a := ValueToPrimitive(test.a)
		b := ValueToPrimitive(test.b)
		c := ValueToPrimitive(test.c)

		got, err := a.Binary(b, test.op)
		if test.err && err == nil {
			t.Errorf("binop should fail: %v %v %v", test.a, test.op, test.b)
		} else if !test.err && err != nil {
			t.Errorf("binop should not fail: %v %v %v (%v)",
				test.a, test.op, test.b, err)
		} else if !reflect.DeepEqual(c, got) {
			t.Errorf("binop expected %v got %v", c, got)
		}
	}
}
