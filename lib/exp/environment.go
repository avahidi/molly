package exp

import (
	"fmt"
	"io"
	"os"

	"bitbucket.org/vahidi/molly/lib/exp/prim"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
	"bitbucket.org/vahidi/molly/lib/util/logging"
)

type env struct {
	// global configuration
	globals *util.Register

	fileSystem types.FileSystem

	// current file
	file io.ReadSeeker

	// current scope
	scope *scope
}

// type assertions
var _ types.Env = (*env)(nil)

func (e env) GetGlobals() *util.Register         { return e.globals }
func (e env) GetFileSystem() types.FileSystem    { return e.fileSystem }
func (e *env) SetFileSystem(fs types.FileSystem) { e.fileSystem = fs }

// ReadSeeker
func (e env) Seek(offset int64, whence int) (int64, error) {
	return e.file.Seek(offset, whence)
}

func (e env) Read(buffer []byte) (int, error) {
	return e.file.Read(buffer)
}

// Stringer
func (e env) String() string {
	if filename, found := e.globals.GetString("$filename", ""); found {
		return fmt.Sprintf("{%s:%s}", e.scope.rule.GetId(), filename)
	}
	return fmt.Sprintf("{%s}", e.scope.rule.GetId())
}

// Scope
func (e *env) Get(id string) (types.Expression, bool) {
	return e.scope.Get(id)
}

func (e *env) Set(id string, exp types.Expression) {
	e.scope.Set(id, exp)
}

func (e env) Extract() map[string]interface{} {
	return e.scope.Extract()
}

func (e env) GetRule() types.Rule {
	return e.scope.GetRule()
}

// Lookup returns a variable either from current scope or
// the global registry
func (e *env) Lookup(id string) (types.Expression, bool, error) {
	exp, found := e.Get(id)

	// lazy evaluation in progress?
	if found && exp == nil {
		logging.Fatalf("Circular dependency in %s (%s)", id, e)
	}
	// attempt resolve lazy evaluation
	if !found {
		if exp, found = e.scope.rule.GetVar(id); found {
			e.scope.Set(id, nil) // show that we are working on it...
			var err error
			exp, err = exp.Eval(e)
			if err != nil {
				return nil, true, err
			}
			e.scope.Set(id, exp)
		}
	}

	if !found {
		var val interface{}
		val, found = e.globals.Get(id)
		if found {
			exp = NewValueExpression(prim.ValueToPrimitive(val))
		}
	}
	return exp, found, nil
}

// state
func (e *env) StartFile(file io.ReadSeeker) {
	e.file = file
	e.Seek(0, os.SEEK_SET)
}

func (e *env) StartRule(rule types.Rule) {
	e.Seek(0, os.SEEK_SET)
	e.scope = newScope(rule, nil)
}

func (e *env) PushScope(newrule types.Rule) {
	e.scope = newScope(newrule, e.scope)
}

func (e *env) PopScope() {
	if e.scope == nil || e.scope.parent == nil {
		fmt.Printf("Internal error: no scope or scope hierarchy")
	}
	e.scope = e.scope.parent
}

// creation
func NewEnvironment(globals *util.Register) types.Env {
	return &env{globals: globals}

}
