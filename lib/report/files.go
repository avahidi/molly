package report

import (
	"bitbucket.org/vahidi/molly/lib/types"
)

// ExtractFileHierarchy creates file hierarchy for generated files
func ExtractFileHierarchy(m *types.Molly) map[string][]string {
	ret := make(map[string][]string)

	for _, fe := range m.Files.Out {
		if fe.Parent != nil {
			tmp := ret[fe.Parent.Filename]
			tmp = append(tmp, fe.Filename)
			ret[fe.Parent.Filename] = tmp
		}
	}
	return ret
}

// ExtractFileList creates a list of all scanned files
func ExtractFileList(m *types.Molly) []string {
	var ret []string
	for _, fe := range m.Files.Out {
		ret = append(ret, fe.Filename)
	}
	return ret
}

// ExtractLogHierarchy creates the  hierarchy for logs generated for files
func ExtractLogHierarchy(r *types.Report) map[string][]string {
	ret := make(map[string][]string)
	for _, file := range r.Files {
		if len(file.Logs) != 0 {
			ret[file.Filename] = file.Logs
		}
	}
	return ret
}

// ExtractFlatReport creates a flat match report
func ExtractFlatReport(r *types.Report) map[string][]*types.FlatMatch {
	ret := make(map[string][]*types.FlatMatch)
	for _, file := range r.Files {
		ret[file.Filename] = ExtractFlatMatches(file)
	}
	return ret
}
