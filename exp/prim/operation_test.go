package prim

import (
	"testing"
	"unicode/utf8"
)

func TestConvert(t *testing.T) {
	var testdata = []struct {
		str string
		op  Operation
	}{
		{"+", ADD},
		{"&&", BAND},
		{"<<", LSL},
		{"!", NEG},
	}

	for _, test := range testdata {
		op := StringToOperation(test.str)
		if op != test.op {
			t.Errorf("Expected %v got %v", test.op, op)
		} else {
			if op.String() != test.str {
				t.Errorf("Expected %v got %v", test.str, op.String())
			}
		}
		if len(test.str) == 1 {
			r, _ := utf8.DecodeRuneInString(test.str)
			op = RuneToOperation(r)
			if op != test.op {
				t.Errorf("Expected %v got %v", test.op, op)
			}
		}
	}
}
