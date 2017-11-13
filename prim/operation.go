package prim

type Operation int

const (
	// arith
	ADD Operation = iota
	SUB
	MUL
	DIV

	// logic and shift/rotate
	AND
	OR
	XOR

	LSL
	LSR

	// compare
	EQ
	NE
	LT
	GT

	// logic
	BAND
	BOR
	BXOR

	// unary
	INV
	NEG
)

var operationNames = [...]string{
	ADD: "+",
	SUB: "-",
	MUL: "*",
	DIV: "/",
	AND: "&",
	OR:  "&",
	XOR: "&",

	LSL: "<<",
	LSR: ">>",

	EQ: "==",
	NE: "!=",
	LT: "<",
	GT: ">",

	BAND: "&&",
	BOR:  "||",
	BXOR: "~",

	INV: "~",
	NEG: "-",
}

func (o Operation) String() string {
	return operationNames[int(o)]
}
