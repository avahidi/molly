package at

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
	case *BinaryExpression:
		Walk(n.Left, v)
		Walk(n.Right, v)
	case *ExtractExpression:
		Walk(n.Size, v)
		Walk(n.Offset, v)
	}
}
