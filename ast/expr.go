package ast

import "github.com/raintreeinc/delphi/token"

type (
	ParenExpr struct {
		Lparen token.Pos
		X      Expr
		Rparen token.Pos
	}

	BinaryExpr struct {
		X     Expr
		OpPos token.Pos
		Op    token.Token
		Y     Expr
	}

	UnaryExpr struct {
		OpPos token.Pos
		Op    token.Token
		X     Expr
	}

	CallExpr struct {
		Name   token.Ident
		Lparen token.Pos
		Args   []Expr
		Rparen token.Pos
	}

	IndexExpr struct {
		X      Expr
		Lbrack token.Pos
		Index  []Expr
		Rbrack token.Pos
	}

	AddrExpr struct {
		At token.Pos // position of '@'
		X  Expr
	}

	DerefExpr struct {
		X   Expr
		Hat token.Pos // position of '^'
	}

	SelectorExpr struct {
		X   Expr
		Sel *Ident
	}

	BasicLit struct {
		ValuePos token.Pos   // literal position
		Kind     token.Token // INTEGER, FLOAT, CHAR, STRING
		Value    string
	}
)
