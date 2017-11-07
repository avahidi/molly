package at

type Operation int

const (
	ADD Operation = iota
	SUB
	MUL
	DIV

	AND
	OR
	XOR

	LSL
	LSR

	EQ
	NE
	LT
	GT

	BAND
	BOR
	BXOR
)
