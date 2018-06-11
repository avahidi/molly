package scan

import (
	"fmt"
	"io"
)

type parser struct {
	// input string
	lex *lexer
}

// Create parser
func newparser(r io.Reader, filename string) *parser {
	return &parser{
		lex: newlexer(r, filename),
	}
}

func (p *parser) next() bool  { return p.lex.next() }
func (p parser) Text() string { return p.lex.text }
func (p parser) Type() rune   { return p.lex.typ }

func (p parser) String() string {
	return fmt.Sprintf("%s", p.lex)
}

// parser helpers
func (p *parser) errorf(format string, args ...interface{}) error {
	return fmt.Errorf("parser: %s (L='%v' T='%s')",
		fmt.Sprintf(format, args...),
		p.lex.scan.Pos(), p.Text(),
	)
}

func (p *parser) accept(t rune, s string) bool {
	if p.Type() == t && p.Text() == s {
		p.next()
		return true
	}
	return false
}

func (p *parser) acceptToken(t rune, str *string) bool {
	if p.Type() == t {
		if str != nil {
			*str = p.Text()
		}
		p.next()
		return true
	}
	return false
}
func (p *parser) acceptTokenAny(ts ...rune) rune {
	for _, t := range ts {
		if p.acceptToken(t, nil) {
			return t
		}
	}
	return None
}

func (p *parser) acceptValue(s string) bool {
	if p.Text() == s {
		p.next()
		return true
	}
	return false
}

func (p *parser) acceptValueAny(typ rune, ss ...string) (bool, string) {
	if p.Type() != typ {
		return false, ""
	}
	for _, s := range ss {
		if p.Text() == s {
			p.next()
			return true, s
		}
	}

	return false, ""
}
