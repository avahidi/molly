package types

// Match represents a rule match on a file
type Match struct {
	Rule *Rule
	Vars map[string]interface{}

	Children []*Match
	Parent   *Match `json:"-"` // this will avoid circular marshalling

	FailedChildren []*Rule `json:"-"` // this will avoid circular marshalling
}

// Walk visits all the nodes in a tree of matches
func (me *Match) Walk(visitor func(*Match) bool) bool {
	if !visitor(me) {
		return false
	}
	for _, c := range me.Children {
		if !c.Walk(visitor) {
			return false
		}
	}
	return true
}

// FlatMatch is a flatten version of Match
type FlatMatch struct {
	Rule *Rule `json:"-"` // dont need this for the reports
	Name string
	Vars map[string]interface{}
}

// Report contains all matches for all files
type Report struct {
	Files []*FileData
}

// NewReport creates a new empty report
func NewReport() *Report {
	return &Report{}
}

// Add adds a FileReport to this report
func (r *Report) Add(fr *FileData) {
	r.Files = append(r.Files, fr)
}

// Add adds a FileReport to this report
func (r Report) RemoveEmpty() *Report {
	r2 := NewReport()
	for _, file := range r.Files {
		if !file.Empty() {
			r2.Add(file)
		}
	}
	return r2
}
