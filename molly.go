package molly

import (
	"io/ioutil"

	_ "github.com/avahidi/molly/operators" // import default actions
	"github.com/avahidi/molly/types"
)

// New creates a new molly context
func New() *types.Molly {
	m := types.NewMolly()

	if m.Config.OutDir == "" {
		m.Config.OutDir, _ = ioutil.TempDir("", "molly-out")
	}
	return m
}

// ExtractReport generates a report
func ExtractReport(m *types.Molly) *types.Report {
	r := &types.Report{}
	for _, file := range m.Files {
		if !file.Empty() {
			r.Files = append(r.Files, file)
		}
	}
	return r
}
