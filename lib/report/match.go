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
func ExtractFlatMatches(fr *types.Input) []*types.FlatMatch {
	var ret []*types.FlatMatch
	for _, match := range fr.Matches {
		fms := ExtractFlatMatch(match)
		ret = append(ret, fms...)
	}
	return ret
}

// FindInReport returns variable from a match
func FindInReport(r *types.Report, filename, rulename, varname string) (interface{}, bool) {
	for _, f := range r.Files {
		if filename != "" && filename != f.Filename {
			continue
		}
		for _, m0 := range f.Matches {
			var match *types.Match
			m0.Walk(func(m *types.Match) bool {
				if rulename == "" || rulename == m.Rule.ID {
					match = m
					return false
				}
				return true
			})
			if match == nil {
				continue
			}
			data, found := match.Vars[varname]
			if found {
				return data, true
			}
		}
	}

	return nil, false
}

// FindInReportNumber is a helper for FindInReport when it returns a number
func FindInReportNumber(r *types.Report, filename, rulename, varname string) (uint64, bool) {
	data, found := FindInReport(r, filename, rulename, varname)
	if !found {
		return 0, false
	}
	num, valid := data.(uint64)
	return num, valid
}

// FindInReportString is a helper for FindInReport when it returns a string
func FindInReportString(r *types.Report, filename, rulename, varname string) (string, bool) {
	data, found := FindInReport(r, filename, rulename, varname)
	if !found {
		return "", false
	}
	str, valid := data.(string)
	return str, valid
}

// FindInReportBool is a helper for FindInReport when it returns a boolean
func FindInReportBool(r *types.Report, filename, rulename, varname string) (bool, bool) {
	data, found := FindInReport(r, filename, rulename, varname)
	if !found {
		return false, false
	}
	b, valid := data.(bool)
	return b, valid
}
