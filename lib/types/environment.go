package types

import (
	"bitbucket.org/vahidi/molly/lib/util"
	"fmt"
	"io"
	"os"
)

// Input represents a file scanned by molly
type Input struct {
	Reader   io.ReadSeeker
	Filename string
	Filesize uint64

	Matches []*Match
	Errors  []error
	Logs    []string
}

// NewInput creates a new Input with given name, size and stream
func NewInput(r io.ReadSeeker, filename string, filesize uint64) *Input {
	return &Input{
		Reader:   r,
		Filename: filename,
		Filesize: filesize,
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

// Env is the current environment during scanning
type Env struct {
	out, log *util.FileSystem
	m        *Molly

	// these are valid while we are scanning a file
	Input *Input
	Scope *Scope
}

func NewEnv(m *Molly) *Env {
	return &Env{
		m:   m,
		out: util.NewFileSystem(m.ExtractDir, m.Files),
		log: util.NewFileSystem(m.ReportDir, nil),
	}
}

func (e *Env) StartRule(rule *Rule) {
	e.Scope = NewScope(rule, nil)
}

func (e *Env) PushRule(newrule *Rule) {
	e.Scope = NewScope(newrule, e.Scope)
}

func (e *Env) PopRule() {
	if e.Scope == nil || e.Scope.Parent == nil {
		util.RegisterFatalf("Internal error: no scope or scope hierarchy")
	}
	e.Scope = e.Scope.Parent
}

func (e Env) String() string {
	if e.Input != nil {
		return fmt.Sprintf("{%s:%s}", e.Scope.Rule.ID, e.Input.Filename)
	}
	return fmt.Sprintf("{%s}", e.Scope.Rule.ID)
}

func (e *Env) SetInput(i *Input) {
	e.Input = i
}

func (e Env) GetFile() string {
	return e.Input.Filename
}

func (e Env) GetSize() uint64 {
	return e.Input.Filesize
}

func (e *Env) Name(name string, addtopath bool) (string, error) {
	return e.out.Name(name, e.GetFile(), addtopath)
}
func (e *Env) Create(name string) (*os.File, error) {
	return e.out.Create(name, e.GetFile())
}

func (e *Env) Mkdir(path string) (string, error) {
	return e.out.Mkdir(path, e.GetFile())
}

// CreateLog creates a new log
func (e *Env) CreateLog(name string) (*os.File, error) {
	file, err := e.log.Create(name, e.GetFile())
	if err == nil && file != nil {
		e.Input.Logs = append(e.Input.Logs, file.Name())
	}
	return file, err
}
