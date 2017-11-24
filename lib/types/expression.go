package types

import (
	"io"

	"bitbucket.org/vahidi/molly/lib/util"
	"bitbucket.org/vahidi/molly/lib/util/logging"
)

// Expression is a node in the AST
type Expression interface {
	Eval(env Env) (Expression, error)
}

// Scope is the current scope for a rule
type Scope interface {
	Get(string) (Expression, bool)
	Set(string, Expression)
	// Delete(string)

	Extract() map[string]interface{}

	GetRule() Rule
}

// Env is the current enviornment during scanning
type Env interface {
	io.ReadSeeker

	GetGlobals() *util.Register
	GetFileSystem() FileSystem
	SetFileSystem(fs FileSystem)

	Scope
	PushScope(Rule)
	PopScope()
	Lookup(id string) (Expression, bool, error)

	StartFile(io.ReadSeeker)
	StartRule(Rule)
}

type Rule interface {
	GetId() string

	GetMetadata() *util.Register
	SetParent(Rule)

	AddCondition(e Expression) error
	AddAction(action Expression) error
	AddVar(id string, e Expression) error
	GetVar(id string) (Expression, bool)

	Close()
	Eval(env Env) (bool, error)

	Actions(func(Rule, Expression) error) error
	Children(func(Rule) error) error
}

// RuleSet represents a group of rules parsed from one or more file
// it also includes the rule hierarchy
type RuleSet struct {
	Files map[string][]Rule
	Top   map[string]Rule
	Flat  map[string]Rule
}

// some helper functions to simplify code elsewhere

// FileName returns the name of the currently scanned binary file
func FileName(e Env) string {
	filename, found := e.GetGlobals().GetString("$filename", "")
	if !found {
		logging.Fatalf("internal error: could not find $filename (%s)", e)
	}
	return filename
}

// FileSize returns the size of the currently scanned binary file
func FileSize(e Env) int64 {
	size, found := e.GetGlobals().GetNumber("$filesize", 0)
	if !found {
		logging.Fatalf("internal error: could not find $filesize (%s)", e)
	}
	return size
}
