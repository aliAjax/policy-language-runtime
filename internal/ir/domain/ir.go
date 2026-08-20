package domain

type Op string

const (
	LoadConst   Op = "const"
	LoadVar     Op = "var"
	Unary       Op = "unary"
	Binary      Op = "binary"
	Call        Op = "call"
	JumpIfFalse Op = "jump_if_false"
	Jump        Op = "jump"
	Return      Op = "return"
	Pop         Op = "pop"
)

type Instr struct {
	Op     Op
	Arg    string
	Value  any
	Target int
}
type Program struct {
	Code   []Instr
	Consts []any
	Deps   []string
}

func (p Program) Clone() Program {
	return p
}
