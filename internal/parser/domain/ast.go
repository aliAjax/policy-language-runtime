package domain

type Expr interface{ expr() }
type Literal struct{ Value any }

func (Literal) expr() {}

type Variable struct{ Name string }

func (Variable) expr() {}

type Unary struct {
	Op    string
	Value Expr
}

func (Unary) expr() {}

type Binary struct {
	Op          string
	Left, Right Expr
}

func (Binary) expr() {}

type Call struct {
	Name string
	Args []Expr
}

func (Call) expr() {}

type Block struct{ Statements []Stmt }
type Stmt interface{ stmt() }
type Let struct {
	Name  string
	Value Expr
}

func (Let) stmt() {}

type Return struct{ Value Expr }

func (Return) stmt() {}

type If struct {
	Cond       Expr
	Then, Else *Block
}

func (If) stmt() {}

type ExprStmt struct{ Value Expr }

func (ExprStmt) stmt() {}

type Program struct{ Statements []Stmt }

func CloneProgram(p *Program) *Program {
	if p == nil {
		return nil
	}
	out := &Program{Statements: make([]Stmt, 0, len(p.Statements))}
	for _, stmt := range p.Statements {
		out.Statements = append(out.Statements, cloneStmt(stmt))
	}
	return out
}

func cloneStmt(stmt Stmt) Stmt {
	switch value := stmt.(type) {
	case Let:
		value.Value = cloneExpr(value.Value)
		return value
	case Return:
		value.Value = cloneExpr(value.Value)
		return value
	case ExprStmt:
		value.Value = cloneExpr(value.Value)
		return value
	case If:
		value.Cond = cloneExpr(value.Cond)
		value.Then = cloneBlock(value.Then)
		value.Else = cloneBlock(value.Else)
		return value
	default:
		return stmt
	}
}

func cloneBlock(block *Block) *Block {
	if block == nil {
		return nil
	}
	out := &Block{Statements: make([]Stmt, 0, len(block.Statements))}
	for _, stmt := range block.Statements {
		out.Statements = append(out.Statements, cloneStmt(stmt))
	}
	return out
}

func cloneExpr(expr Expr) Expr {
	switch value := expr.(type) {
	case Unary:
		value.Value = cloneExpr(value.Value)
		return value
	case Binary:
		value.Left = cloneExpr(value.Left)
		value.Right = cloneExpr(value.Right)
		return value
	case Call:
		value.Args = append([]Expr(nil), value.Args...)
		for i := range value.Args {
			value.Args[i] = cloneExpr(value.Args[i])
		}
		return value
	default:
		return expr
	}
}
