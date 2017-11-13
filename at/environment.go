package at

import (
	"fmt"
	"io"
	"os"
	"path"

	"bitbucket.org/vahidi/molly/prim"
)

type Scope struct {
	Variables     map[string]Expression
	DefaultFormat prim.Format
}

func NewScope() *Scope {
	return &Scope{
		Variables: make(map[string]Expression),
	}
}

func (s Scope) Get(id string) Expression {
	return s.Variables[id]
}

func (s *Scope) Set(id string, exp Expression) {
	s.Variables[id] = exp
}

func (e Scope) Dump() {
	for k, v := range e.Variables {
		fmt.Printf("%s = %s\n", k, v)
	}
}

type Env struct {
	file  io.ReadSeeker
	Scope *Scope
}

func NewEnvironmentFromFile(file io.ReadSeeker, size uint64, filename string) *Env {
	e := &Env{file: file, Scope: NewScope()}
	e.file.Seek(0, os.SEEK_SET)

	dir, name := path.Dir(filename), path.Base(filename)
	e.Scope.Variables["base$"] = NewStringExpression(path.Join(dir, "DATA_"+name))
	e.Scope.Variables["filename$"] = NewStringExpression(name)
	e.Scope.Variables["filesize$"] = NewNumberExpression(size, 8, prim.UBE)
	return e
}

func (e Env) Dump() {
	fmt.Printf("Enviorment %v\n", e)
	e.Scope.Dump()
}
