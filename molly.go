package lib

import (
	"io/ioutil"

	_ "bitbucket.org/vahidi/molly/actions" // import default actions
	"bitbucket.org/vahidi/molly/types"
)

// New creates a new molly context
func New() *types.Molly {
	m := types.NewMolly()

	if m.Config.OutDir == "" {
		m.Config.OutDir, _ = ioutil.TempDir("", "molly-out")
	}
	return m
}
