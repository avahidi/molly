package prim

import "unicode/utf8"

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
	NEG: "!",
}

func (o Operation) String() string {
	return operationNames[int(o)]
}

var runeToOperationMap map[rune]Operation
var stringToOperationMap map[string]Operation

func RuneToOperation(r rune) Operation     { return runeToOperationMap[r] }
func StringToOperation(s string) Operation { return stringToOperationMap[s] }

func init() {
	runeToOperationMap = make(map[rune]Operation)
	stringToOperationMap = make(map[string]Operation)
	for i, s := range operationNames {
		stringToOperationMap[s] = Operation(i)
		if len(s) == 1 {
			r, _ := utf8.DecodeRuneInString(s)
			runeToOperationMap[r] = Operation(i)
		}
	}
}
