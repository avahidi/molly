package types

import (
	"fmt"
	"io"
	"os"

	"bitbucket.org/vahidi/molly/lib/util/logging"

	"bitbucket.org/vahidi/molly/lib/util"
)

// Env is the current enviornment during scanning
type Env struct {
	Globals    *util.Register
	FileSystem FileSystem

	// these are valid while we are scanning a file
	file  io.ReadSeeker
	Scope *Scope
}

func NewEnv() *Env {
	return &Env{Globals: util.NewRegister()}
}

// ReadSeeker
func (e Env) Seek(offset int64, whence int) (int64, error) {
	return e.file.Seek(offset, whence)
}

func (e Env) Read(buffer []byte) (int, error) {
	return e.file.Read(buffer)
}

func (e *Env) StartFile(file io.ReadSeeker) {
	e.file = file
	e.Seek(0, os.SEEK_SET)
}

func (e *Env) StartRule(rule *Rule) {
	e.Seek(0, os.SEEK_SET)
	e.Scope = NewScope(rule, nil)
}

func (e *Env) PushScope(newrule *Rule) {
	e.Scope = NewScope(newrule, e.Scope)
}

func (e *Env) PopScope() {
	if e.Scope == nil || e.Scope.Parent == nil {
		logging.Fatalf("Internal error: no scope or scope hierarchy")
	}
	e.Scope = e.Scope.Parent
}

func (e Env) String() string {
	if filename, found := e.Globals.GetString("$filename", ""); found {
		return fmt.Sprintf("{%s:%s}", e.Scope.Rule.ID, filename)
	}
	return fmt.Sprintf("{%s}", e.Scope.Rule.ID)
}

/*

type Env interface {
	io.ReadSeeker

	GetGlobals() *util.Register
	GetFileSystem() FileSystem
	SetFileSystem(fs FileSystem)

	// Scope
	GetScope() *Scope
	PushScope(*Rule)
	PopScope()
	Lookup(id string) (Expression, bool, error)

	StartFile(io.ReadSeeker)
	StartRule(*Rule)
}
*/
