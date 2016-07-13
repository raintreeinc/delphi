package ast

import "go/token"

type Type interface {
	Node
}

type ArrayType struct {
	Start token.Pos // position of "array" keyword
	Dim   []ArrayTypeDim
	Type  Type
}

type ArrayTypeDim struct {
	Low, High Expr
}

type SetType struct {
	Start token.Pos // position of "set" keyword
	Type  Type
}

type PointerType struct {
	Start token.Pos // position of caret ^
	Type  Type
}

type NamedType struct {
	Ident Ident
}
