package at

import (
	"fmt"
	"io"
	"os"
	"path"
)

type Env struct {
	AddFile func(filename string)
	NewFile func(suggestedName string) string

	// current file
	file io.ReadSeeker
	// current class
	scope       *Scope
	globalScope *Scope
}

// ReadSeeker
func (e Env) Seek(position int64) error {
	_, err := e.file.Seek(position, os.SEEK_SET)
	return err
}

func (e Env) Read(buffer []byte) (int, error) {
	return e.file.Read(buffer)
}

// Scope
func (e Env) Get(id string) (Expression, bool) {
	exp, found := e.scope.Get(id)
	if !found {
		exp, found = e.globalScope.Get(id)
	}
	return exp, found
}

func (e *Env) Set(id string, exp Expression) {
	e.scope.Set(id, exp)
}

func (e Env) Extract() map[string]interface{} {
	return e.scope.Extract()
}

// state
func (e *Env) SetFile(file io.ReadSeeker, size uint64, filename string) {
	e.file = file
	e.Seek(0)
	dir, name := path.Dir(filename), path.Base(filename)

	e.globalScope.Set("path$", NewStringExpression(dir))
	e.globalScope.Set("filename$", NewStringExpression(name))
	e.globalScope.Set("filesize$", NewNumberExpression(size, 8, false))
}

func (e *Env) Reset() {
	e.Seek(0)
	e.scope = NewScope()
}

func (e *Env) PushScope() {
	s := NewScope()
	s.parent = e.scope
	e.scope = s
}

func (e *Env) PopScope() {
	if e.scope == nil || e.scope.parent == nil {
		fmt.Printf("Internal error: no scope or scope hierarchy")
	}
	e.scope = e.scope.parent
}

// Debug
func (e Env) Dump() {
	fmt.Printf("Enviorment %v\n", e)
	e.scope.Dump()
}

// creation
func NewEnvironment(
	newFile func(suggestedName string) string,
	addFile func(filename string)) *Env {
	return &Env{
		AddFile:     addFile,
		NewFile:     newFile,
		globalScope: NewScope(),
	}
}
