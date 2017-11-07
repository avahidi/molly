package at

import (
	"fmt"
	"io"
	"os"
)

type Scope struct {
	Variables     map[string]Expression
	DefaultFormat Format
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

func NewEnviorment(file io.ReadSeeker) *Env {
	e := &Env{file: file}
	e.Reset()
	return e
}

func (e *Env) Reset() {
	e.Scope = NewScope()
	e.file.Seek(0, os.SEEK_SET)
}

func (e Env) Dump() {
	fmt.Printf("Enviorment %v\n", e)
	e.Scope.Dump()
}
