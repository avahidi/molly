package types

import (
	"bitbucket.org/vahidi/molly/lib/util"
)

// Molly represents the context of a molly program
type Molly struct {
	ExtractDir string
	ReportDir  string

	Rules *RuleSet
	Files *util.FileQueue

	OnMatchRule func(file *Input, match *Match)
	OnMatchTag  func(file *Input, tag string)
}

// NewMolly creates a new Molly context
func NewMolly(extratDir, reportDir string) *Molly {
	return &Molly{
		ExtractDir: extratDir,
		ReportDir:  reportDir,
		Rules:      NewRuleSet(),
		Files:      util.NewFileQueue(),
	}
}
