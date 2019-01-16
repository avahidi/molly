package types

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Analysis represents results of an analysis performed on a file
type Analysis struct {
	Name   string
	Result interface{}
	Error  error
}

type FileData struct {
	Parent *FileData

	Filename    string
	FilenameOut string
	Filesize    int64

	time time.Time

	Checksum []byte

	// hierarchy
	Depth       int
	Children    []*FileData
	DuplicateOf *FileData

	// These are filled as we scan the file
	Processed bool
	Matches   []*Match
	Errors    []error
	Warnings  []string
	Logs      []string
	Analyses  map[string]*Analysis
}

func NewFileData(filename string, parent *FileData) *FileData {
	fd := &FileData{
		Filename:    filename,
		FilenameOut: filename,
		Parent:      parent,
		time:        time.Now(),
		Analyses:    make(map[string]*Analysis),
	}
	// update parent data and make sure child is not newer than parent
	if parent != nil {
		fd.Depth = parent.Depth + 1
		fd.time = parent.time
	}
	return fd
}

func (fd *FileData) SetTime(t time.Time) {
	fd.time = t
	if fd.Parent != nil && t.After(fd.Parent.time) {
		fd.time = fd.Parent.time
	}
}

func (fd *FileData) GetTime() time.Time {
	return fd.time
}

func (fd FileData) Empty() bool {
	return len(fd.Matches) == 0 && len(fd.Errors) == 0 && len(fd.Logs) == 0
}

// RegisterErrorf registers an error
func (fd *FileData) RegisterErrorf(format string, v ...interface{}) {
	fd.RegisterError(fmt.Errorf(format, v...))
}

// RegisterError registers an error
func (fd *FileData) RegisterError(err error) {
	fd.Errors = append(fd.Errors, err)
}

// RegisterWarning registers a warning
func (fd *FileData) RegisterWarning(format string, v ...interface{}) {
	fd.Warnings = append(fd.Warnings, (fmt.Sprintf(format, v...)))
}

func (fd *FileData) RegisterAnalysis(name string, data interface{}, err error) {
	fd.Analyses[name] = &Analysis{Name: name, Result: data, Error: err}
}

// Get returns variables associated with this file.
// These can be referensed in rules as $name or
// in the actions as {name}
func (fd FileData) Get(name string) (interface{}, bool) {
	// note: if you update this one, also update FileDataGetHelp
	switch name {
	case "time":
		return fd.time, true
	case "filename":
		return fd.Filename, true
	case "shortname":
		_, shortname := filepath.Split(fd.Filename)
		return shortname, true
	case "dirname":
		dirname, _ := filepath.Split(fd.Filename)
		return dirname, true
	case "ext":
		return filepath.Ext(fd.Filename), true
	case "basename":
		base := filepath.Base(fd.Filename)
		n := strings.LastIndex(base, ".")
		if n == -1 {
			return base, true
		}
		return base[:n], true
	case "filesize":
		return fd.Filesize, true
	case "depth":
		return fd.Depth, true
	case "parent":
		if fd.Parent == nil {
			return "", true
		}
		return fd.Parent.Filename, true
	case "num_matches":
		return len(fd.Matches), true
	case "num_errors":
		return len(fd.Errors), true
	case "num_logs":
		return len(fd.Logs), true
	default:
		return nil, false
	}
}

// FileDataGetHelp dump help about the special variables such as $time
func FileDataGetHelp() {
	list := []string{
		"time",
		"filename", "shortname", "dirname", "ext", "basename",
		"filesize", "depth", "parent",
		"num_matches", "num_errors", "num_logs",
	}
	fmt.Printf("Valid special variables are:\n")
	for _, v := range list {
		fmt.Printf("\t%s\n", v)
	}
}
