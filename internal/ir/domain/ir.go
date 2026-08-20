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
	p.Code = append([]Instr(nil), p.Code...)
	for i := range p.Code {
		switch value := p.Code[i].Value.(type) {
		case []string:
			p.Code[i].Value = append([]string(nil), value...)
		case []any:
			p.Code[i].Value = append([]any(nil), value...)
		}
	}
	p.Consts = append([]any(nil), p.Consts...)
	p.Deps = append([]string(nil), p.Deps...)
	return p
}
