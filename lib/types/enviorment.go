package types

import (
	"bitbucket.org/vahidi/molly/lib/util"
	"fmt"
	"io"
	"path/filepath"
)

// Env is the current enviornment during scanning
type Env struct {
	Globals    *util.Register

	FileSystem FileSystem

	// these are valid while we are scanning a file
	Reader io.ReadSeeker
	Scope  *Scope
}

func NewEnv() *Env {
	return &Env{
		Globals: util.NewRegister(),
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
	if filename, found := e.Globals.GetString("$filename", ""); found {
		return fmt.Sprintf("{%s:%s}", e.Scope.Rule.ID, filename)
	}
	return fmt.Sprintf("{%s}", e.Scope.Rule.ID)
}

func (e *Env) SetFile(filename string, filesize uint64) {
	path, name := filepath.Split(filename)
	e.Globals.SetString("$path", path)
	e.Globals.SetString("$shortfilename", name)
	e.Globals.SetString("$filename", filename)
	e.Globals.SetNumber("$filesize", filesize)
}

func (e *Env) GetFilesystem() FileSystem {
	return nil //
}