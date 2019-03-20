package types

import (
	"fmt"
	"io"
	"os"

	"bitbucket.org/vahidi/molly/util"
)

// Env is the current environment during scanning
type Env struct {
	m *Molly

	// Input is valid while we are scanning a file
	Reader  io.ReadSeeker
	Current *FileData

	// Scope is valid while we are scanning a file and a rule
	Scope *Scope
}

func NewEnv(m *Molly) *Env {
	return &Env{
		m: m,
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
	if e.Current != nil {
		return fmt.Sprintf("{%s:%s}", e.Scope.Rule.ID, e.Current.Filename)
	}
	return fmt.Sprintf("{%s}", e.Scope.Rule.ID)
}

func (e *Env) SetInput(r io.ReadSeeker, d *FileData) {
	e.Reader = r
	e.Current = d
}

func (e Env) GetFile() string {
	return e.Current.Filename
}

func (e Env) GetSize() uint64 {
	return uint64(e.Current.Filesize)
}

func (e *Env) New(name string, islog bool) (*FileData, error) {
	return e.m.New(e.Current, name, false, islog)
}

func (e *Env) Create(name string) (*os.File, *FileData, error) {
	return e.m.CreateFile(e.Current, name, false)
}

func (e *Env) Mkdir(path string) (*FileData, error) {
	return e.m.CreateDir(e.Current, path)
}

// CreateLog creates a new log
func (e *Env) CreateLog(name string) (*os.File, error) {
	newfile, _, err := e.m.CreateFile(e.Current, name, true)
	return newfile, err
}

func (e Env) HasPermission(p Permission) bool {
	return e.m.Config.HasPermission(p)
}
