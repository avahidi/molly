package parser

import (
	"fmt"
	"os"
	"text/scanner"

	"bitbucket.org/vahidi/molly/at"
	"bitbucket.org/vahidi/molly/prim"
	"bitbucket.org/vahidi/molly/util"
)

var (
	ErrorUnexpected = fmt.Errorf("Unexpected token")
)

// RuleSet represents a group of rules parsed from one or more file
// it also includes the rule hierarchy
type RuleSet struct {
	Files      map[string][]*at.Rule
	Rules      []*at.Rule
	map_       map[string]*at.Rule
	parentship map[string]string
}

func newRuleSet() *RuleSet {
	return &RuleSet{
		Files:      make(map[string][]*at.Rule),
		map_:       make(map[string]*at.Rule),
		parentship: make(map[string]string),
	}
}

func ParseRules(files []string) (*RuleSet, error) {
	rs := newRuleSet()

	for _, filename := range files {
		var list []*at.Rule
		r, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		// fmt.Printf("Scaning rule %s...\n", filename)

		p := NewParser(r)
		p.Next()
		for {
			if p.AcceptToken(scanner.EOF, nil) {
				break
			}
			c, parent, err := parseRule(p)
			if err != nil {
				fmt.Println("ERROR", err)
				return nil, err
			}
			list = append(list, c)
			rs.map_[c.Id] = c
			if parent != "" {
				rs.parentship[c.Id] = parent
			}
		}
		rs.Files[filename] = list
	}

	// from the flattened rs.map_ and parent relationship rs.parnetship build
	// the rule hierarchy:
	for _, klass := range rs.map_ {
		parent, has := rs.parentship[klass.Id]
		if has {
			p, has := rs.map_[parent]
			if !has {
				return nil, fmt.Errorf("Rule %s is subRule of non-existing %s", klass.Id, parent)
			}
			p.Children = append(p.Children, klass)
			klass.Parent = p
		} else {
			rs.Rules = append(rs.Rules, klass)
		}
	}
	return rs, nil
}

func parseRule(p *Parser) (c *at.Rule, parent string, e error) {
	if !p.AcceptValue("rule") {
		return nil, "", ErrorUnexpected
	}

	id := ""
	if !p.AcceptToken(scanner.Ident, &id) {
		e = ErrorUnexpected
		return
	}

	parent = ""
	if p.AcceptToken(':', nil) {
		if !p.AcceptToken(scanner.Ident, &parent) {
			e = ErrorUnexpected
			return
		}
	}
	if !p.AcceptToken('{', nil) {
		e = ErrorUnexpected
		return
	}

	c = at.NewRule(id)
	for {

		if p.AcceptToken('}', nil) {
			return
		} else if p.AcceptValue("var") {
			e = parseAssignment(p, c)
		} else if p.AcceptValue("if") {
			e = parseCondition(p, c)
		} else {
			e = parseAction(p, c)
		}

		if e != nil {
			return
		}

		// end of statement
		if !p.AcceptToken(';', nil) {
			e = ErrorUnexpected
			return
		}
	}
}

func parseAssignment(p *Parser, c *at.Rule) error {
	var id string
	if !p.AcceptToken(scanner.Ident, &id) {
		return ErrorUnexpected
	}

	if !p.Accept(Operator, "=") {
		return ErrorUnexpected
	}
	expr, err := parseExpression(p)
	if err == nil {
		c.AddAssignment(id, expr)
	}
	return err
}

func parseCondition(p *Parser, c *at.Rule) error {
	expr, err := parseExpression(p)
	if err == nil {
		c.AddCondition(expr)
	}
	return err
}

func parseAction(p *Parser, c *at.Rule) error {
	var name string
	if !p.AcceptToken(scanner.Ident, &name) {
		return ErrorUnexpected
	}
	expr, err := parseCall(p, name)
	if err != nil {
		return err
	}
	c.AddAction(expr)
	return nil
}

func parseExpression(p *Parser) (at.Expression, error) {
	return parseBinary(p)
}

func parseBinary(p *Parser) (at.Expression, error) {
	u1, err := parseUnary(p)
	if err != nil {
		return nil, err
	}

	got, op := p.AcceptValueAny(Operator, "-", "+", "*", "/",
		"==", "!=", ">", "<",
		"&", "&&", "|", "||", "^")
	if !got {
		return u1, nil
	}

	u2, err := parseUnary(p)
	if err != nil {
		return nil, err
	}
	return at.NewBinaryExpression(u1, u2, prim.StringToOperation(op)), nil
}

func parseUnary(p *Parser) (at.Expression, error) {
	got, op := p.AcceptValueAny(Operator, "-", "+", "~", "!")
	e, err := parsePrimary(p)
	if err != nil {
		return nil, err
	}
	if got {
		return at.NewUnaryExpression(e, prim.StringToOperation(op)), nil
	}
	return e, nil
}

func parsePrimary(p *Parser) (at.Expression, error) {
	var str string

	// string
	if p.AcceptToken(scanner.String, &str) {
		str := str[1 : len(str)-1] // remove quotes
		ss := prim.NewString(str)
		return at.NewValueExpression(ss), nil
	}

	// number?
	if p.AcceptToken(scanner.Int, &str) {
		n, err := util.ParseNumber(str, 64)
		if err != nil {
			return nil, err
		}
		num := prim.NewNumber(n, 4, false)
		return at.NewValueExpression(num), nil
	}

	// identifier?
	if p.AcceptToken(scanner.Ident, &str) {
		// sepcial cases?
		if str == "true" {
			return at.NewValueExpression(prim.NewBoolean(true)), nil
		}
		if str == "false" {
			return at.NewValueExpression(prim.NewBoolean(false)), nil
		}

		// function call or a variable access?
		if p.Type() == '(' {
			return parseCall(p, str)
		} else {
			return &at.VariableExpression{Id: str}, nil
		}
	}

	if p.AcceptToken('(', nil) {
		expr, err := parseExpression(p)
		if err != nil {
			return nil, err
		}
		if !p.AcceptToken(')', nil) {
			return nil, ErrorUnexpected
		}
		return expr, nil
	}

	// byte array
	if p.AcceptToken('{', nil) {
		vals := make([]byte, 0)
		for {
			if p.AcceptToken(scanner.Int, &str) {
				bb, err := util.ParseNumber(str, 8)
				if err != nil {
					return nil, err
				}
				vals = append(vals, byte(bb))
			} else if p.AcceptToken(scanner.Char, &str) {
				vals = append(vals, byte(str[1]))
			} else {
				return nil, p.Errorf("Unexpected item in byte array")
			}
			if p.AcceptToken('}', nil) {
				break
			}
			if !p.AcceptToken(',', nil) {
				return nil, ErrorUnexpected
			}
		}
		bytes := prim.NewStringRaw(vals)
		return at.NewValueExpression(bytes), nil

	}

	return nil, p.Errorf("unknown expression")
}

func parseCall(p *Parser, id string) (at.Expression, error) {

	if !p.AcceptToken('(', nil) {
		return nil, ErrorUnexpected
	}
	argv := make([]at.Expression, 0)
	for {
		e, err := parseExpression(p)
		if err != nil {
			return nil, err
		}
		argv = append(argv, e)

		if p.AcceptToken(')', nil) {
			break
		}
		if !p.AcceptToken(',', nil) {
			return nil, ErrorUnexpected
		}
	}

	argc := len(argv)

	switch id {
	case "StringZ":
		if argc == 2 {
			return at.NewExtractExpression(argv[0], argv[1], at.StringZ), nil
		}
	case "String":
		if argc == 2 {
			return at.NewExtractExpression(argv[0], argv[1], at.String), nil
		}
	case "Byte":
		if argc == 1 {
			return at.NewExtractExpression(argv[0],
				at.NewNumberExpression(1, 4, false), at.NumberBEU), nil
		}
	case "Short":
		if argc == 1 {
			return at.NewExtractExpression(argv[0],
				at.NewNumberExpression(2, 4, false), at.NumberBEU), nil
		}
	case "Long":
		if argc == 1 {
			return at.NewExtractExpression(argv[0],
				at.NewNumberExpression(4, 4, false), at.NumberBEU), nil
		}
	case "Quad":
		if argc == 1 {
			return at.NewExtractExpression(argv[0],
				at.NewNumberExpression(8, 4, false), at.NumberBEU), nil
		}
	default:
		return at.NewFunctionExpression(id, argv...)
	}

	return nil, fmt.Errorf("incorrect arguments: %s %s", id, argv)
}
