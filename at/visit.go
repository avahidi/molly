package at

/*
// Visitor can visit a node
type Visitor interface {
	Visit(Expression) Visitor
}

// Walk can visit every node in the tree
func Walk(e Expression, v Visitor) {
	if v = v.Visit(e); v == nil {
		return
	}
	switch n := e.(type) {
	case *OperationExpression:
		Walk(n.Left, v)
		if n.Right != nil {
			Walk(n.Right, v)
		}
	case *ExtractExpression:
		Walk(n.Size, v)
		Walk(n.Offset, v)
	}
}

*/
