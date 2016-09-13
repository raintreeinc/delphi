package ast

import "github.com/raintreeinc/delphi/token"

type (
	Node interface {
	}

	Decl interface {
		Node
	}
	Expr interface {
		Node
	}

	Define struct {
		Not  bool
		Name string
	}

	Ident struct {
		NamePos token.Pos // identifier position
		Name    string    // identifier name
		Obj     *Object   // denoted object or nil
	}

	Unit struct {
		Doc   *CommentGroup
		Start token.Pos // position of "unit" keyword
		Name  Ident

		Iface Section
		Impl  Section
	}

	Section struct {
		Start token.Pos // position of "interface" or "implementation" keyword
		Decl  []Decl
	}

	Class struct {
		Doc       *CommentGroup
		Name      *Ident
		Ancestors []Ident
		Scopes    []QualifiedDecls
	}

	QualifiedDecls struct {
		Doc       *CommentGroup
		Start     token.Pos   // position of Qualifier
		Qualifier token.Token // one of PRIVATE, PROTECTED, PUBLIC, PUBLISHED, RECORD
		Decls     []Decl
	}

	FieldList struct {
		Names []Decl
		Type  Type
	}

	Property struct {
		Start   token.Pos // position of "property" keyword
		Name    Ident
		Type    Type
		Array   *ArgumentList
		Index   Expr
		Read    *Ident
		Write   *Ident
		Stored  Expr
		Default Expr
	}

	// functions/procedures and methods

	FuncDecl struct {
		Start token.Pos   // position of Token
		Token token.Token // one of FUNCTION, PROCEDURE, DESTRUCTOR, CONSTRUCTOR, INITIALIZATION, FINALIZATION

		Doc        *CommentGroup   // comments immediately before the func
		Recv       Type            // receiver, if specified
		Name       *Ident          // function/method/procedure name, if specified
		Args       []ArgumentList  // arguments for the function
		Result     Type            // result identifier
		Directives []FuncDirective // directives specified for this function
		Body       *FuncBody       // when one exists
	}

	FuncDirective struct {
		Start token.Pos
		Token token.Token // one of VIRTUAL, ABSTRACT, OVERRIDE, STDCALL, EXTERNAL, ...
		Param interface{}
	}

	ArgumentList struct {
		Kind    token.Token // either token.INVALID or token.VAR, token.CONST, token.OUT
		Names   []Ident     // list of names specified
		Type    Type        // can be nil
		Default Expr        // can be nil
	}

	// Declaring of variables and constants

	Vars struct {
		Doc   *CommentGroup
		Start token.Pos // position of "var"
		List  []*Var
	}
	Consts struct {
		Doc   *CommentGroup
		Start token.Pos // position of "const"
		List  []*Var
	}

	Var struct {
		Doc     *CommentGroup
		Name    *Ident
		Type    Type
		Default Expr
	}

	// Comments

	CommentGroup struct {
		List []*Comment
	}

	Comment struct {
		Start token.Pos
		End   token.Pos
		Text  string
	}
)
