package scan

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/scanner"
)

type Parser struct {
	// input string
	lex *Lexer
}

func (p *Parser) Next()        { p.lex.Next() }
func (p *Parser) Text() string { return p.lex.Text() }
func (p *Parser) Token() rune  { return p.lex.Token }

/*
func (p *Parser) Rest() string {
	off := p.lex.scan.Pos().Offset - len(p.Text())
	p.lex.Token = scanner.EOF // hack to force EOF
	return p.input[off:]
}
*/
// Create parser
func NewParser(r io.Reader) *Parser {
	return &Parser{
		lex: NewLexer(r),
	}
}

// parser helpers
func (p *Parser) Errorf(format string, args ...interface{}) error {
	return fmt.Errorf("Parser: %s (L='%s' %=%s)",
		fmt.Sprintf(format, args...),
		p.lex.scan.Pos(), p.Text())
}

func (p *Parser) Accept(t rune) (bool, string) {
	if p.Token() == t {
		s := p.Text()
		p.Next()
		return true, s
	}
	return false, ""
}

func (p *Parser) AcceptAny(ts ...rune) (rune, int) {
	for i, t := range ts {
		if a, _ := p.Accept(t); a {
			return t, i
		}
	}
	return 0, -1
}

func (p *Parser) RejectAny(ts ...rune) (rune, int) {
	for i, t := range ts {
		if p.Token() == t {
			return t, i
		}
	}
	p.Next()
	return 0, -1
}

func (p *Parser) AcceptValue(s string) bool {
	if p.Text() == s {
		p.Next()
		return true
	}
	return false
}

func (p *Parser) AcceptAnyValue(vs ...string) (string, int) {
	for i, v := range vs {
		if p.Text() == v {
			p.Next()
			return v, i
		}
	}
	return "", -1
}

func (p *Parser) RequireUnsigned(bits int) (uint64, error) {
	a, num := p.Accept(scanner.Int)
	if !a {
		return 0, p.Errorf("expected number")
	}

	if strings.HasPrefix(num, "0x") || strings.HasPrefix(num, "0X") {
		return strconv.ParseUint(num[2:], 16, bits)
	}
	return strconv.ParseUint(num, 10, bits)
}
