package report

import (
	"github.com/avahidi/molly/types"
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
	m.Walk(func(match *types.Match) bool {
		if len(match.Children) != 0 {
			return true // we will takes its children instead
		}
		var flat = &types.FlatMatch{
			Rule: match.Rule,
			Name: match.Rule.ID,
			Vars: make(map[string]interface{}),
		}
		flattenMatch(match, flat)
		ret = append(ret, flat)
		return true
	})

	return ret
}

// ExtractFlatMatches creates a flat match list for all rules that have no children
func ExtractFlatMatches(fr *types.FileData) []*types.FlatMatch {
	var ret []*types.FlatMatch
	for _, match := range fr.Matches {
		fms := ExtractFlatMatch(match)
		ret = append(ret, fms...)
	}
	return ret
}

// ExtractMatchNames return name of all matches
func ExtractMatchNames(i *types.FileData, flatten bool) []string {
	var ret []string
	for _, m := range i.Matches {
		m.Walk(func(match *types.Match) bool {
			if !flatten && len(match.Children) > 0 {
				return true
			}
			ret = append(ret, match.Rule.ID)
			return true
		})
	}
	return ret
}
