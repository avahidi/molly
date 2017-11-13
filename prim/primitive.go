package prim

type Primitive interface {
	Get() interface{}
	Extract(data []byte) error
	Binary(o Primitive, op Operation) (Primitive, error)
	Unary(op Operation) (Primitive, error)
}
