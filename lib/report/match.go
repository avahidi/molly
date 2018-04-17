package report

import (
	"bitbucket.org/vahidi/molly/lib/types"
)

// flattenMatch extracts variables from a match entry and its parents.
// Note that if this entry has children they will be ignored
func flattenMatch(in *types.Match, out *types.FlatMatch) {
	if in.Parent != nil {
		flattenMatch(in.Parent, out)
	}
	for k, v := range in.Vars {
		out.Vars[k] = v
	}
}

// ExtractFlatMatch flattens one match hierarchy
func ExtractFlatMatch(m *types.Match) []*types.FlatMatch {
	var ret []*types.FlatMatch
	m.Walk(func(match *types.Match) {
		if len(match.Children) != 0 {
			return // we will takes its children instead
		}
		var flat = &types.FlatMatch{
			Rule: match.Rule,
			Name: match.Rule.ID,
			Vars: make(map[string]interface{}),
		}
		flattenMatch(match, flat)
		ret = append(ret, flat)
	})

	return ret
}

// ExtractFlatMatches creates a flat match list for all rules that have no children
func ExtractFlatMatches(fr *types.Input) []*types.FlatMatch {
	var ret []*types.FlatMatch
	for _, match := range fr.Matches {
		fms := ExtractFlatMatch(match)
		ret = append(ret, fms...)
	}
	return ret
}
