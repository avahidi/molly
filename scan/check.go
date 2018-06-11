package scan

import (
	"fmt"
	"reflect"

	"bitbucket.org/vahidi/molly/exp"
	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
)

// constraint defines valid (meta) data
type constraint map[string]struct {
	kind reflect.Kind
	ctrl func(name string, data interface{}) error
}

// checkMetadata checks a metadata register against some constraints
func checkMetadata(r *util.Register, cs constraint) error {
	var err error
	r.Walk(func(v string, d interface{}) bool {
		c, found := cs[v]
		if !found {
			err = fmt.Errorf("Unknown metadata: %s", v)
			return false
		}
		typ := reflect.TypeOf(d)
		k := typ.Kind()
		if k != c.kind {
			err = fmt.Errorf("Metadata %s: expected type '%v' got '%v", v, c.kind, k)
			return false
		}
		if c.ctrl != nil {
			err = c.ctrl(v, d)
			if err != nil {
				return false
			}
		}
		return true
	})
	return err
}

// checkRule controls if a rule has any errors so far undetected
func checkRule(r *types.Rule) error {
	var cs = constraint{
		"tag":       {reflect.String, nil},
		"bigendian": {reflect.Bool, nil},
		"pass": {reflect.Uint64, func(name string, data interface{}) error {
			if r.Parent != nil {
				return fmt.Errorf("Only parent rules can have 'pass'")
			}
			n := data.(uint64)
			if n < types.RulePassMin || n > types.RulePassMax {
				return fmt.Errorf("valid range for pass is %d to %d",
					types.RulePassMin, types.RulePassMax)
			}
			return nil
		}},
	}

	return checkMetadata(r.Metadata, cs)
}

// checkFunction controls if a function call has any errors
func checkFunction(f *exp.FunctionExpression) error {
	var cs = constraint{
		"bigendian": {reflect.Bool, nil},
	}

	return checkMetadata(f.Metadata, cs)
}
