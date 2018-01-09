package types

import (
	"fmt"
	"io"

	"bitbucket.org/vahidi/molly/lib/util"
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
	return &Env{Globals: util.NewRegister()}
}

func (e *Env) StartRule(rule *Rule) {
	e.Scope = NewScope(rule, nil)
}

func (e *Env) PushScope(newrule *Rule) {
	e.Scope = NewScope(newrule, e.Scope)
}

func (e *Env) PopScope() {
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
