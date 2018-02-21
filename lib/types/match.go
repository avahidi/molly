package types

// Match represents a rule match on a file
type Match struct {
	Rule *Rule
	Vars map[string]interface{}

	Children []*Match
	Parent   *Match `json:"-"` // this will avoid circular marshallingw

	FailedChildren []*Rule `json:"-"` // this will avoid circular marshalling
}

// Walk visits all the nodes in a tree of matches
func (me *Match) Walk(visitor func(*Match)) {
	visitor(me)
	for _, c := range me.Children {
		c.Walk(visitor)
	}
}

// FlatMatch is a flatten version of Match
type FlatMatch struct {
	Rule *Rule `json:"-"` // dont need this for the reports
	Name string
	Vars map[string]interface{}
}

// FileReport contains all matches + some otherdata for one file
type FileReport struct {
	Filename string
	Matches  []*Match
	Errors   []error
	Logs     []string
}

// NewFileReport creates a new empty report for a file
func NewFileReport(filename string) *FileReport {
	return &FileReport{Filename: filename}
}

// Empty returns true if this report contains no data
func (fr FileReport) Empty() bool {
	return len(fr.Matches) == 0 && len(fr.Errors) == 0 && len(fr.Logs) == 0
}

// Walk walks the file report for all matches including hierarchical matches
func (fr FileReport) Walk(visitor func(*Match)) {
	for _, match := range fr.Matches {
		match.Walk(visitor)
	}
}

// Report contains all matches for all files
type Report struct {
	Current *FileReport
	Files   []*FileReport
}

// NewReport creates a new empty report
func NewReport() *Report {
	return &Report{}
}

// Add adds a FileReport to this report
func (mr *Report) Add(fr *FileReport) {
	mr.Current = fr
	mr.Files = append(mr.Files, fr)
}
