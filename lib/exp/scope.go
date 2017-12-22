package exp

import "bitbucket.org/vahidi/molly/lib/types"

// Extract values from all expressions that have any
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
