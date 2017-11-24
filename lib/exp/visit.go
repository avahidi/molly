package exp

import "bitbucket.org/vahidi/molly/lib/types"

// visitor can visit a node in an expression
type visitor func(types.Expression) visitor

// walk can visit every node in the tree
func walk(e types.Expression, v visitor) {
	if e == nil {
		return
	}

	if v = v(e); v == nil {
		return
	}
	switch n := e.(type) {
	case *OperationExpression:
		walk(n.Left, v)
		walk(n.Right, v)
	case *ExtractExpression:
		walk(n.Size, v)
		walk(n.Offset, v)
	case *SliceExpression:
		walk(n.Expr, v)
		walk(n.Start, v)
		walk(n.End, v)
	case *FunctionExpression:
		for _, param := range n.Params {
			walk(param, v)
		}
	}

}
