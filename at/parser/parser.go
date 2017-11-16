package parser

import (
	"fmt"
	"io"
)

type Parser struct {
	// input string
	lex *Lexer
}

// Create parser
func NewParser(r io.Reader) *Parser {
	return &Parser{
		lex: NewLexer(r),
	}
}

func (p *Parser) Next() bool  { return p.lex.Next() }
func (p Parser) Text() string { return p.lex.Text() }
func (p Parser) Type() rune   { return p.lex.Type() }

func (p Parser) String() string {
	return fmt.Sprintf("%s", p.lex)
}

// parser helpers
func (p *Parser) Errorf(format string, args ...interface{}) error {
	return fmt.Errorf("Parser: %s (L='%s' %=%s)",
		fmt.Sprintf(format, args...),
		p.lex.scan.Pos(), p.Text())
}

func (p *Parser) Accept(t rune, s string) bool {
	if p.Type() == t && p.Text() == s {
		p.Next()
		return true
	}
	return false
}

func (p *Parser) AcceptToken(t rune, str *string) bool {
	if p.Type() == t {
		if str != nil {
			*str = p.Text()
		}
		p.Next()
		return true
	}
	return false
}
func (p *Parser) AcceptTokenAny(ts ...rune) rune {
	for _, t := range ts {
		if p.AcceptToken(t, nil) {
			return t
		}
	}
	return None
}

func (p *Parser) AcceptValue(s string) bool {
	if p.Text() == s {
		p.Next()
		return true
	}
	return false
}

func (p *Parser) AcceptValueAny(typ rune, ss ...string) (bool, string) {
	if p.Type() != typ {
		return false, ""
	}
	for _, s := range ss {
		if p.Text() == s {
			p.Next()
			return true, s
		}
	}

	return false, ""
}
