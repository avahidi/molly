package operators

import (
	"bitbucket.org/vahidi/molly/types"
)

// OperatorHelp prints help text including signature for all registred actions
func Help() {
	types.OperatorHelp()

	ChecksumHelp()
	AnalyzerHelp()
	ExtractorHelp()
}

func Register(name string, fun interface{}) error {
	return types.OperatorRegister(name, fun)
}
