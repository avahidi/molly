package scan

import (
	"fmt"
	"os"
	"text/scanner"

	"bitbucket.org/vahidi/molly/lib/exp"
	"bitbucket.org/vahidi/molly/lib/exp/prim"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
	"bitbucket.org/vahidi/molly/lib/util/logging"
)

func precedence(op string) int {
	switch op {
	case "*", "/", "%":
		return 7
	case "+", "-":
		return 6
	case "<<", ">>":
		return 5
	case "<", "<=", ">", ">=":
		return 4
	case "==", "!=":
		return 3
	case "&&", "&":
		return 2
	case "||", "|", "^":
		return 1
	}
	return 0
}

func ParseRules(ins types.FileQueue, rs *types.RuleSet) error {
	var listAll []*types.Rule
	parents := make(map[string]string)

	for filename := ins.Pop(); filename != ""; filename = ins.Pop() {
		listFile, _ := rs.Files[filename]
		r, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer r.Close()

		// parse rules in a file and make sure they haven't been seen before
		p := newparser(r, filename)
		p.next()
		for !p.acceptToken(scanner.EOF, nil) {
			c, parent, err := parseRule(p)
			if err != nil {
				return err
			}
			id := c.ID
			if _, found := rs.Flat[id]; found {
				return fmt.Errorf("Multiple definitions of '%s' (last in %s)", id, filename)
			}
			listFile = append(listFile, c)
			listAll = append(listAll, c)
			rs.Flat[id] = c
			if parent != "" {
				parents[id] = parent
			}
		}
		rs.Files[filename] = listFile
	}

	// from the flattened rs.map_ and parent relationship rs.parnetship build
	// the rule hierarchy:
	for _, klass := range listAll {
		kid := klass.ID
		parent, has := parents[kid]
		if has {
			p, has := rs.Flat[parent]
			if !has {
				return fmt.Errorf("Rule %s is missing parent %s", kid, parent)
			}

			// p is now parent to klass
			klass.Parent = p
			p.Children = append(p.Children, klass)
			klass.Metadata.SetParent(p.Metadata)
		} else {
			rs.Top[kid] = klass
		}
	}
	// now close all rules
	for _, klass := range listAll {
		exp.RuleClose(klass)
	}
	return nil
}

func parseRule(p *parser) (*types.Rule, string, error) {
	if !p.acceptValue("rule") {
		return nil, "", p.errorf("Unknown token, expected rule")
	}

	var id string
	if !p.acceptToken(scanner.Ident, &id) {
		return nil, "", p.errorf("Unknown token, expected rule identifier")
	}
	c := types.NewRule(id)

	// chec if we have rule metadata
	if p.acceptToken('(', nil) {

		for {
			var metaid string
			if !p.acceptToken(scanner.Ident, &metaid) {
				return nil, "", p.errorf("Expected metadata identifier")
			}
			if metaid[0] == '$' {
				return nil, "", p.errorf("Invalid identifier")
			}
			if !p.acceptValue("=") {
				return nil, "", p.errorf("Expected '=' in metadata")
			}

			val, err := parseConstant(p)
			if err != nil {
				return nil, "", err
			}
			c.Metadata.Set(metaid, val)

			if p.acceptToken(')', nil) {
				break
			}
			if !p.acceptToken(',', nil) {
				return nil, "", p.errorf("Expected ',' in rule metadata")
			}
		}
	}

	// check if we have a parent
	parent := ""
	if p.acceptToken(':', nil) {
		if !p.acceptToken(scanner.Ident, &parent) {
			return nil, "", p.errorf("Unknown token, expected rule parent identifier")
		}
	}
	if !p.acceptToken('{', nil) {
		return nil, "", p.errorf("Unknown token, expected {")
	}

	// parse rule components
	for {
		var e error
		if p.acceptToken('}', nil) {
			return c, parent, nil
		} else if p.acceptValue("var") {
			e = parseAssignment(p, c)
		} else if p.acceptValue("if") {
			e = parseCondition(p, c)
		} else {
			e = parseAction(p, c)
		}

		if e != nil {
			return nil, "", e
		}

		// end of statement
		if !p.acceptToken(';', nil) {
			return nil, "", p.errorf("Unknown token, expected ';'")
		}
	}
}

func parseAssignment(p *parser, c *types.Rule) error {
	var id string
	if !p.acceptToken(scanner.Ident, &id) {
		return p.errorf("Unknown token, expected LHS in assignment")
	}
	if id[0] == '$' {
		return p.errorf("Invalid identifier")
	}

	if !p.accept(Operator, "=") {
		return p.errorf("Unknown token, expected = for assignment")
	}
	expr, err := parseExpression(p)
	if err == nil {
		if _, exists := c.Variables[id]; exists {
			err = fmt.Errorf("variable '%s' is already defined in %s", id, c.ID)
		} else {
			c.Variables[id] = expr
		}
	}
	return err
}

func parseCondition(p *parser, c *types.Rule) error {
	expr, err := parseExpression(p)
	if err != nil {
		return err
	}
	c.Conditions = append(c.Conditions, exp.Simplify(expr))
	return nil
}

func parseAction(p *parser, c *types.Rule) error {
	var name string
	if !p.acceptToken(scanner.Ident, &name) {
		return p.errorf("Unknown token, expected function name in action")
	}
	expr, err := parseCall(p, name)
	if err != nil {
		return err
	}
	c.Actions = append(c.Actions, exp.Simplify(expr))
	return nil
}

func parseExpression(p *parser) (types.Expression, error) {
	return parseBinary(p, 1)
}

func parseBinary(p *parser, maxPrec int) (types.Expression, error) {
	u1, err := parseUnary(p)
	if err != nil {
		return nil, err
	}

	for prec := precedence(p.Text()); p.Type() == Operator && prec >= maxPrec; prec-- {

		for p.Type() == Operator {
			op := p.Text()
			if precedence(op) != prec {
				break
			}
			p.next()
			u2, err := parseBinary(p, prec+1)
			if err != nil {
				return nil, err
			}
			u1 = exp.NewBinaryExpression(u1, u2, prim.StringToOperation(op))
		}
	}
	return u1, nil
}

func parseUnary(p *parser) (types.Expression, error) {
	got, op := p.acceptValueAny(Operator, "-", "+", "~", "!")
	e, err := parsePrimary(p)
	if err != nil {
		return nil, err
	}
	if got {
		return exp.NewUnaryExpression(e, prim.StringToOperation(op)), nil
	}
	return e, nil
}

func parsePrimary(p *parser) (types.Expression, error) {
	var str string

	// string
	if p.acceptToken(scanner.String, &str) {
		str := str[1 : len(str)-1] // remove quotes
		// ss := prim.NewString(str)
		bs, err := util.DecodeString([]byte(str))
		if err != nil {
			return nil, err
		}
		ss := prim.NewStringRaw(bs)
		return exp.NewValueExpression(ss), nil
	}

	// number?
	if p.acceptToken(scanner.Int, &str) {
		n, err := util.ParseNumber(str, 64)
		if err != nil {
			return nil, err
		}
		num := prim.NewNumber(n, 8, true)
		return exp.NewValueExpression(num), nil
	}

	// identifier?
	if p.acceptToken(scanner.Ident, &str) {
		// sepcial cases?
		if str == "true" {
			return exp.NewValueExpression(prim.NewBoolean(true)), nil
		}
		if str == "false" {
			return exp.NewValueExpression(prim.NewBoolean(false)), nil
		}

		// function call or a variable access?
		if p.Type() == '(' {
			return parseCall(p, str)
		} else {
			var err error = nil
			var v types.Expression = &exp.VariableExpression{Id: str}

			// is it a slice
			if p.acceptToken('[', nil) {
				v, err = parseSlice(p, v)
			}
			return v, err
		}
	}

	if p.acceptToken('(', nil) {
		expr, err := parseExpression(p)
		if err != nil {
			return nil, err
		}
		if !p.acceptToken(')', nil) {
			return nil, p.errorf("Unknown token, expected ')'")
		}
		return expr, nil
	}

	// byte array
	if p.acceptToken('{', nil) {
		vals := make([]byte, 0)
		for {
			if p.acceptToken(scanner.Int, &str) {
				bb, err := util.ParseNumber(str, 8)
				if err != nil {
					return nil, err
				}
				vals = append(vals, byte(bb))
			} else if p.acceptToken(scanner.Char, &str) {
				vals = append(vals, byte(str[1]))
			} else {
				return nil, p.errorf("Unexpected item in byte array")
			}
			if p.acceptToken('}', nil) {
				break
			}
			if !p.acceptToken(',', nil) {
				return nil, p.errorf("Unknown token, expected ',' in array ")
			}
		}
		bytes := prim.NewStringRaw(vals)
		return exp.NewValueExpression(bytes), nil

	}

	return nil, p.errorf("unknown expression")
}

func parseSlice(p *parser, v types.Expression) (types.Expression, error) {
	start, err := parseExpression(p)
	if err != nil {
		return nil, err
	}

	var end types.Expression
	if p.acceptToken(':', nil) {
		end, err = parseExpression(p)
		if err != nil {
			return nil, err
		}
	}
	if !p.acceptToken(']', nil) {
		return nil, p.errorf("Bad index, expected ']")
	}

	return exp.NewSliceExpression(v, start, end), nil
}

func parseConstant(p *parser) (interface{}, error) {
	e2, err := parsePrimary(p)
	if err != nil {
		return nil, err
	}
	e2 = exp.Simplify(e2)
	val, valid := e2.(*exp.ValueExpression)
	if !valid {
		return nil, fmt.Errorf("Expected constant value: '%v'", e2)
	}

	return val.Value.Get(), err
}

func parseCall(p *parser, id string) (types.Expression, error) {
	if !p.acceptToken('(', nil) {
		return nil, p.errorf("Unknown token, expected '(' in function call ")
	}

	metadata := util.NewRegister()
	argv := make([]types.Expression, 0)

	for {
		e1, err := parseExpression(p)
		if err != nil {
			return nil, err
		}

		// special case for parsing f(..., stuff = value)
		// stuff is an identifier, so currently it is stored as VariableExpression
		if p.acceptValue("=") {
			e10, valid := e1.(*exp.VariableExpression)
			if !valid {
				return nil, fmt.Errorf("Expected identifer in metadata, got '%v'", e1)
			}
			id := e10.Id

			val, err := parseConstant(p)
			if err != nil {
				return nil, err
			}
			metadata.Set(id, val)
			fmt.Printf("Metadata: %s = %v (%T)\n", id, val, val)
		} else {
			argv = append(argv, e1)
		}

		if p.acceptToken(')', nil) {
			break
		}

		if !p.acceptToken(',', nil) {
			return nil, p.errorf("Expected ',' in call parameters")
		}
	}

	expr, err := findExtractFunction(id, argv, metadata)
	if err == nil && expr == nil {
		// not an extract function? try a regular one
		if _, found := types.FunctionFind(id); !found {
			fmt.Printf("Unknown function '%s'. ", id)
			types.FunctionHelp()
			logging.Fatalf("Unknown function, cannot continue")
		}

		expr, err = exp.NewFunctionExpression(id, metadata, argv...)
	}

	return expr, err
}

// findExtractFunction figures out if this is an extract function and
// returns the corrector extractor for it
func findExtractFunction(id string, argv []types.Expression, metadata *util.Register) (
	types.Expression, error) {

	argc := len(argv)
	format := exp.ExtractFormat{}

	switch id {
	case "StringZ":
		if argc == 2 {
			format.Type = exp.StringZ
			return exp.NewExtractExpression(argv[0], argv[1], metadata, format), nil
		}
	case "String":
		if argc == 2 {
			format.Type = exp.String
			return exp.NewExtractExpression(argv[0], argv[1], metadata, format), nil
		}
	case "Byte":
		if argc == 1 {
			format.Type = exp.UNumber
			return exp.NewExtractExpression(argv[0],
				exp.NewNumberExpression(1, 4, false), metadata, format), nil
		}
	case "Short":
		if argc == 1 {
			format.Type = exp.UNumber
			return exp.NewExtractExpression(argv[0],
				exp.NewNumberExpression(2, 4, false), metadata, format), nil
		}
	case "Long":
		if argc == 1 {
			format.Type = exp.UNumber
			return exp.NewExtractExpression(argv[0],
				exp.NewNumberExpression(4, 4, false), metadata, format), nil
		}
	case "Quad":
		if argc == 1 {
			format.Type = exp.UNumber
			return exp.NewExtractExpression(argv[0],
				exp.NewNumberExpression(8, 4, false), metadata, format), nil
		}
	default:
		// no error but neither an extract function
		return nil, nil
	}

	return nil, fmt.Errorf("incorrect arguments: %s %s", id, argv)
}
