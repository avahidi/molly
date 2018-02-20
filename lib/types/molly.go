package types


// Molly represents the context of a molly program
type Molly struct {
	ExtractDir string
	ReportDir string

	Rules *RuleSet
}