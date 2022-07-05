package scan

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

type lexer struct {
	scan scanner.Scanner
	typ  rune
	text string
}

func (l *lexer) second(if_ rune, then_ rune, else_ rune) bool {
	if l.scan.Peek() == if_ {
		l.scan.Scan()
		l.text = l.text + l.scan.TokenText()
		l.typ = then_
		return true
	}
	l.typ = else_
	return false
}

func (l *lexer) next() bool {
	l.typ = l.scan.Scan()
	l.text = l.scan.TokenText()

	switch l.typ {
	case '+', '-', '*', '/', '^', '%', '~':
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

	case '$':
		// identifiers may start with $
		if !l.next() || l.typ != Ident {
			return false
		}
		l.text = fmt.Sprintf("$%s", l.text)

	default:
		/* no changes needed */
	}

	return l.typ != EOF
}

func (l lexer) String() string {
	return fmt.Sprintf("Token{'%v' '%s'}", l.typ, l.text)
}

func newlexer(r io.Reader, filename string) *lexer {
	s := scanner.Scanner{}
	s.Init(r)
	s.Filename = filename
	// s.Mode = mode

	ret := &lexer{scan: s}
	return ret
}
