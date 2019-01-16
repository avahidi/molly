package molly

import (
	"strings"

	"bitbucket.org/vahidi/molly/scan"
	"bitbucket.org/vahidi/molly/types"
)

// LoadRules reads rules from files
func LoadRules(m *types.Molly, files ...string) error {
	return scan.ParseRuleFiles(m, files...)
}

// LoadRulesFromText reads rules from a string
func LoadRulesFromText(m *types.Molly, text string) error {
	return scan.ParseRuleStream(m, strings.NewReader(text))
}
