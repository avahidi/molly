package types

// Expression is a node in the AST
type Expression interface {
	Eval(env *Env) (Expression, error)
	Simplify() (Expression, error)
}
