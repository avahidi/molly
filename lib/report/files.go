package report

import (
	"bitbucket.org/vahidi/molly/lib/types"
)

// ExtractFileHierarchy creates file hierarchy for generated files
func ExtractFileHierarchy(m *types.Molly) map[string][]string {
	ret := make(map[string][]string)
	for _, i := range m.Processed {
		if i.Parent != nil {
			tmp := ret[i.Parent.Filename]
			tmp = append(tmp, i.Filename)
			ret[i.Parent.Filename] = tmp
		}
	}
	return ret
}

// ExtractFileList creates a list of all scanned files
func ExtractFileList(m *types.Molly) []string {
	var ret []string
	for _, i := range m.Processed {
		ret = append(ret, i.Filename)
	}
	return ret
}

// ExtractLogHierarchy creates the  hierarchy for logs generated for files
func ExtractLogHierarchy(r *types.Report) map[string][]string {
	ret := make(map[string][]string)
	for _, i := range r.Files {
		if len(i.Logs) != 0 {
			ret[i.Filename] = i.Logs
		}
	}
	return ret
}
