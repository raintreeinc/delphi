package ast

import "github.com/raintreeinc/delphi/token"

type (
	Type interface {
		Node
	}

	ArrayType struct {
		Start token.Pos // position of "array" keyword
		Dim   []ArrayTypeDim
		Type  Type
	}

	ArrayTypeDim struct {
		Low, High Expr
	}

	SetType struct {
		Start token.Pos // position of "set" keyword
		Type  Type
	}

	PointerType struct {
		Start token.Pos // position of caret ^
		Type  Type
	}

	NamedType struct {
		Ident Ident
	}
)
