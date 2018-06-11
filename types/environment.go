package types

import (
	"fmt"
	"os"

	"bitbucket.org/vahidi/molly/util"
)

// Env is the current environment during scanning
type Env struct {
	m *Molly

	// Input is valid while we are scanning a file
	Input *Input

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
	return uint64(e.Input.Filesize)
}

func (e *Env) Name(name string, islog bool) (string, error) {
	return e.m.CreateName(e.Input, name, false, islog), nil
}

func (e *Env) Create(name string) (*os.File, error) {
	return e.m.CreateFile(e.Input, name, false)
}

func (e *Env) Mkdir(path string) (string, error) {
	return e.m.CreateDir(e.Input, path)
}

// CreateLog creates a new log
func (e *Env) CreateLog(name string) (*os.File, error) {
	return e.m.CreateFile(e.Input, name, true)
}
