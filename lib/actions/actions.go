package actions

import (
	"bitbucket.org/vahidi/molly/lib/types"
)

// ActionHelp prints help text including signature for all registred actions
func ActionHelp() {
	types.FunctionHelp()

	ChecksumHelp()
	AnalyzerHelp()
	ExtractorHelp()
}

func ActionRegister(name string, fun interface{}) error {
	return types.FunctionRegister(name, fun)
}
