package report

import "bitbucket.org/vahidi/molly/types"

// FindInReportFile find a file in a report, filename "" means any
func FindInReportFile(r *types.Report, filename string) *types.Input {
	for _, f := range r.Files {
		if filename == "" || filename == f.Filename {
			return f
		}
	}
	return nil
}

// FindInReportMatch find a file in a report, filename or rulename "" means any
func FindInReportMatch(r *types.Report, filename, rulename string) *types.Match {
	for _, f := range r.Files {
		if filename != "" && filename != f.Filename {
			continue
		}
		if m := FindInFileMatch(f, rulename); m != nil {
			return m
		}
	}
	return nil
}

// FindInFileMatch find match to a rule in a file, rulename "" means any
func FindInFileMatch(f *types.Input, rulename string) *types.Match {
	var match *types.Match
	for _, m0 := range f.Matches {
		m0.Walk(func(m *types.Match) bool {
			if rulename == "" || rulename == m.Rule.ID {
				match = m
				return false
			}
			return true
		})
		if match != nil {
			return match
		}
	}
	return nil
}

// FindInFileVar finds a variable in a match in a file
func FindInFileVar(f *types.Input, rulename, varname string) (interface{}, bool) {
	match := FindInFileMatch(f, rulename)
	if match == nil {
		return nil, false
	}
	data, found := match.Vars[varname]
	return data, found
}

// FindInReportVar returns variable from a match
func FindInReportVar(r *types.Report, filename, rulename, varname string) (interface{}, bool) {
	for _, f := range r.Files {
		if filename != "" && filename != f.Filename {
			continue
		}
		if data, found := FindInFileVar(f, rulename, varname); found {
			return data, true
		}
	}

	return nil, false
}

// FindInReportVarNumber is a helper for FindInReport when it returns a number
func FindInReportVarNumber(r *types.Report, filename, rulename, varname string) (uint64, bool) {
	data, found := FindInReportVar(r, filename, rulename, varname)
	if !found {
		return 0, false
	}
	num, valid := data.(uint64)
	return num, valid
}

// FindInReportVarString is a helper for FindInReport when it returns a string
func FindInReportVarString(r *types.Report, filename, rulename, varname string) (string, bool) {
	data, found := FindInReportVar(r, filename, rulename, varname)
	if !found {
		return "", false
	}
	str, valid := data.(string)
	return str, valid
}

// FindInReportVarBool is a helper for FindInReport when it returns a boolean
func FindInReportVarBool(r *types.Report, filename, rulename, varname string) (bool, bool) {
	data, found := FindInReportVar(r, filename, rulename, varname)
	if !found {
		return false, false
	}
	b, valid := data.(bool)
	return b, valid
}
