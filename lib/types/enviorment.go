package types

import (
	"bitbucket.org/vahidi/molly/lib/util"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Env is the current enviornment during scanning
type Env struct {
	out, log *util.FileSystem
	Globals  *util.Register

	// these are valid while we are scanning a file
	Reader io.ReadSeeker
	Scope  *Scope
}

func NewEnv(config *Config, queue *util.FileQueue) *Env {
	// in case we dont have a functioning config, create one
	if config == nil {
		config = &Config{}
	}
	if config.ExtractionDir == "" {
		config.ExtractionDir, _ = ioutil.TempDir("", "molly-out")
	}

	if config.LogDir == "" {
		config.LogDir, _ = ioutil.TempDir("", "molly-log")
	}
	return &Env{
		out:     util.NewFileSystem(config.ExtractionDir, queue),
		log:     util.NewFileSystem(config.LogDir, nil),
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

func (e Env) GetFile() string {
	filename, _ := e.Globals.GetString("$filename", "")
	return filename
}

func (e Env) GetSize() uint64 {
	size, _ := e.Globals.GetNumber("$filesize", 0)
	return size
}

// Output returns the output filesystem
func (e Env) Output() *util.FileSystem { return e.out }

func (e *Env) Name(name string, addtopath bool) (string, error) {
	return e.out.Name(name, e.GetFile(), addtopath)
}
func (e *Env) Create(name string) (*os.File, error) {
	return e.out.Create(name, e.GetFile())
}

func (e *Env) Mkdir(path string) (string, error) {
	return e.out.Mkdir(path, e.GetFile())
}

// Log returns the log filesystem
func (e Env) Log() *util.FileSystem { return e.log }

// CreateLog creates a new log
func (e *Env) CreateLog(name string) (*os.File, error) {
	return e.log.Create(name, e.GetFile())
}
