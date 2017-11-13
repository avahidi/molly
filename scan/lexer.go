package scan

import (
	"io"
	"text/scanner"
)

type Lexer struct {
	scan  scanner.Scanner
	Token rune
}

func (l *Lexer) Next()        { l.Token = l.scan.Scan() }
func (l *Lexer) Text() string { return l.scan.TokenText() }

func NewLexer(r io.Reader) *Lexer {
	s := scanner.Scanner{}
	s.Init(r)
	// s.Mode = mode

	ret := &Lexer{scan: s}
	ret.Next()
	return ret
}
