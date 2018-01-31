package types

// MatchEntry represents a rule match on a file
type MatchEntry struct {
	Filename string
	Rule     string
	Vars     map[string]interface{}

	FailedChildren []string
	Children       []*MatchEntry
	Parent         *MatchEntry `json:"-"` // this will avoid circular marshalling
}

// Walk visits all the nodes in a tree of matches
func (me *MatchEntry) Walk(visitor func(*MatchEntry)) {
	visitor(me)
	for _, c := range me.Children {
		c.Walk(visitor)
	}
}

// MatchReport contains the results of scanning a number of files
type MatchReport struct {
	Files        []string
	OutHierarchy map[string][]string
	LogHierarchy map[string][]string
	Tagged       map[string][]string
	Errors       []error
	MatchTree    []*MatchEntry
}

// NewMatchReport creates a new MatchReport
func NewMatchReport() *MatchReport {
	return &MatchReport{
		OutHierarchy: make(map[string][]string),
		LogHierarchy: make(map[string][]string),
		Tagged:       make(map[string][]string),
	}
}
