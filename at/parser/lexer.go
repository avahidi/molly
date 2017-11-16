package parser

import (
	"fmt"
	"io"
	"text/scanner"
)

const (
	EOF       = scanner.EOF
	Ident     = scanner.Ident
	Int       = scanner.Int
	Float     = scanner.Float
	Char      = scanner.Char
	String    = scanner.String
	RawString = scanner.RawString
	Comment   = scanner.Comment
)
const (
	Operator rune = -(iota + 10)
	None          // not scanned, only used by accept-any and similar
)

type Lexer struct {
	scan scanner.Scanner
	typ  rune
	text string
}

func (l Lexer) Text() string { return l.text }
func (l Lexer) Type() rune   { return l.typ }

func (l *Lexer) second(if_ rune, then_ rune, else_ rune) bool {
	if l.scan.Peek() == if_ {
		l.scan.Scan()
		l.text = l.text + l.scan.TokenText()
		l.typ = then_
	}
	l.typ = else_
	return false
}

func (l *Lexer) Next() bool {
	l.typ = l.scan.Scan()
	l.text = l.scan.TokenText()

	switch l.typ {
	case '+', '-', '*', '/', '^':
		l.typ = Operator
	case '&':
		l.second('&', Operator, Operator)
	case '|':
		l.second('|', Operator, Operator)
	case '<':
		if !l.second('=', Operator, Operator) {
			l.second('<', Operator, Operator)
		}
	case '>':
		if !l.second('=', Operator, Operator) {
			l.second('>', Operator, Operator)
		}
	case '=':
		l.second('=', Operator, Operator)
	case '!':
		l.second('=', Operator, Operator)

	case Ident:
		// we allow identifiers to end with $?
		l.second('$', Ident, Ident)

	default:
		/* no changes needed */
	}

	return l.typ != EOF
}

func (l Lexer) String() string {
	return fmt.Sprintf("Token{'%v' '%s'}", l.typ, l.text)
}

func NewLexer(r io.Reader) *Lexer {
	s := scanner.Scanner{}
	s.Init(r)
	// s.Mode = mode

	ret := &Lexer{scan: s}
	return ret
}
