package exp

import "github.com/avahidi/molly/types"

// ScopeExtract extracts value from a scope
func ScopeExtract(s *types.Scope) map[string]interface{} {
	ret := make(map[string]interface{})
	for k, v := range s.GetAll() {
		if _, found := ret[k]; !found {
			if vv, fin := v.(*ValueExpression); fin {
				ret[k] = vv.Value.Get()
			}
		}
	}
	return ret
}
