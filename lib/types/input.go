package types

import (
	"io"
	"path/filepath"
	"strings"
)

// Input represents a file scanned by molly
type Input struct {
	Reader      io.ReadSeeker
	Filename    string
	FilenameOut string
	Filesize    int64

	// hierarchy
	Depth    int
	Parent   *Input
	Children []string

	// These are filled as we scan the file
	Matches []*Match
	Errors  []error
	Logs    []string
}

// NewInput creates a new Input with given name, size and stream
func NewInput(filename string, filesize int64) *Input {
	return &Input{
		Filename:    filename,
		FilenameOut: filename,
		Filesize:    filesize,
	}
}

// Read Implements io.Reader
func (i *Input) Read(p []byte) (n int, err error) {
	return i.Reader.Read(p)
}

// Seek Implements io.Seeker
func (i *Input) Seek(offset int64, whence int) (int64, error) {
	return i.Reader.Seek(offset, whence)
}

// Empty returns true if this report contains no data
func (i Input) Empty() bool {
	return len(i.Matches) == 0 && len(i.Errors) == 0 && len(i.Logs) == 0
}

// Get returns variables associated with this file.
// These can be referensed in rules as $name or
// in the actions as {name}
func (i Input) Get(name string) (interface{}, bool) {
	switch name {
	case "filename":
		return i.Filename, true
	case "shortname":
		_, shortname := filepath.Split(i.Filename)
		return shortname, true
	case "dirname":
		dirname, _ := filepath.Split(i.Filename)
		return dirname, true
	case "ext":
		return filepath.Ext(i.Filename), true
	case "basename":
		base := filepath.Base(i.Filename)
		n := strings.LastIndex(base, ".")
		if n == -1 {
			return base, true
		}
		return base[:n], true
	case "filesize":
		return i.Filesize, true
	case "depth":
		return i.Depth, true
	case "parent":
		if i.Parent == nil {
			return "", true
		}
		return i.Parent.Filename, true
	case "num_matches":
		return len(i.Matches), true
	case "num_errors":
		return len(i.Errors), true
	case "num_logs":
		return len(i.Logs), true
	default:
		return nil, false
	}
}
