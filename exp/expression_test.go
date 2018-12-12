package exp

import (
	"testing"

	"bitbucket.org/vahidi/molly/exp/prim"
	"bitbucket.org/vahidi/molly/types"
)

func TestSimplify(t *testing.T) {
	var err error
	var a, b, c, d types.Expression
	a = NewNumberExpression(100, 4, true)
	b = NewNumberExpression(51, 4, true)
	c = NewBinaryExpression(a, b, prim.ADD)
	d = NewUnaryExpression(a, prim.NEG)

	a, err = a.Simplify()
	if err != nil {
		t.Errorf("Simplify a failed: %v", err)
	}
	b, err = b.Simplify()
	if err != nil {
		t.Errorf("Simplify b failed: %v", err)
	}

	c, err = c.Simplify()
	if err != nil {
		t.Errorf("Simplify c failed: %v", err)
	}

	cv, valid := c.(*ValueExpression)
	if !valid || cv.Value.Get() != int32(151) {
		t.Errorf("Should have simplified %v", c)
	}

	d, err = d.Simplify()
	if err != nil {
		t.Errorf("Simplify d failed: %v", err)
	}

	dv, valid := d.(*ValueExpression)
	if !valid || dv.Value.Get() != int32(-100) {
		t.Errorf("Should have simplified %v", d)
	}
}
