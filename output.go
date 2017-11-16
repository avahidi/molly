package molly

// MatchEntry represents a rule match on a file
type MatchEntry struct {
	Filename string
	Rule     string
	Vars     map[string]interface{}
}

// ScanSet contains the results of scanning a number of files
type OutputSet struct {
	Results []*MatchEntry
	Errors  []error
}
