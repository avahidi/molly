package util

import (
	"testing"
)

// fsimple simple function used for testing Function
func fsimple(sum, n1, n2 int) (int, bool) {
	return n1 + n2, sum == n1+n2
}

// fvariadic is a variadic function used for testing Function
func fvariadic(sum int, values ...int) (int, bool) {
	realsum := 0
	for _, v := range values {
		realsum += v
	}
	return realsum, sum == realsum
}

func TestBadFunction(t *testing.T) {
	badfuncs := []interface{}{"joe", false, uint32(123)}
	for _, bf := range badfuncs {
		if _, err := NewFunction("bad", bf); err == nil {
			t.Errorf("Accepted non-function '%v' as function", bf)
		}
	}
}

func TestSimpleFunction(t *testing.T) {
	f1, err := NewFunction("simpel", fsimple)
	if err != nil {
		t.Errorf("Simple call creation failed: '%v' ", err)
	}

	e0, e1 := fsimple(1, 2, 3)
	ret, err := f1.Call([]interface{}{1, 2, 3})

	if err != nil {
		t.Errorf("Simple call failed: '%v' ", err)
	}

	if len(ret) != 2 {
		t.Errorf("Simple call wrong return values. expected 2 got %d ", len(ret))
	}

	r0, okay := ret[0].(int)
	if !okay || r0 != e0 {
		t.Errorf("Simple call first return value incorrect: %d %v\n", ret[0], okay)
	}

	r1, okay := ret[1].(bool)
	if !okay || r1 != e1 {
		t.Errorf("Simple call second return value incorrect: %d %v\n", ret[1], okay)
	}

}

func TestVariadicFunction(t *testing.T) {
	f1, err := NewFunction("variadic", fvariadic)
	if err != nil {
		t.Errorf("Variadic call creation failed: '%v' ", err)
	}

	e0, e1 := fvariadic(1, 2, 3)
	ret, err := f1.Call([]interface{}{1, 2, 3})

	if err != nil {
		t.Errorf("Variadic call failed: '%v' ", err)
	}

	if len(ret) != 2 {
		t.Errorf("Variadic call wrong return values. expected 2 got %d ", len(ret))
	}

	r0, okay := ret[0].(int)
	if !okay || r0 != e0 {
		t.Errorf("Variadic call first return value incorrect: %t %v (expected %d)\n", ret[0], okay, e0)
	}

	r1, okay := ret[1].(bool)
	if !okay || r1 != e1 {
		t.Errorf("Variadic call second return value incorrect: %t %v (expected %v)\n", ret[1], okay, e1)
	}
}

func TestBadVariadicFunction(t *testing.T) {
	f1, err := NewFunction("variadic", fvariadic)
	if err != nil {
		t.Errorf("Variadic call creation failed: '%v' ", err)
	}

	_, err = f1.Call([]interface{}{})
	if err == nil {
		t.Errorf("Bad variadic call should have failed")
	}

	_, err = f1.Call([]interface{}{1})
	if err != nil {
		t.Errorf("Empty variadic should not have failed: %v", err)
	}
}
