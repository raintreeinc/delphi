// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Adapted for Delphi

// Package token defines constants representing the lexical tokens of Delphi
// language and basic operations on tokens (printing, predicates).
//
package token

import (
	"strconv"
	"strings"
)

// Token is the set of lexical tokens of the Go programming language.
type Token int

// The list of tokens.
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	COMMENT    // {} (* *) //
	CDIRECTIVE // {$I HELLO} {IFDEF X}

	literal_beg
	// Identifiers and basic type literals
	// (these tokens stand for classes of literals)
	IDENT
	INTEGER // 12, $12A
	FLOAT   // 0.412, -1e6
	CHAR    // #13, #$1A, ^C
	STRING  // 'x', 'hello world'
	literal_end

	operator_beg
	// Operators and delimiters
	AT  // @
	HAT // ^

	ADD  // +
	SUB  // -
	MUL  // *
	FDIV // /

	EQL // =
	LSS // <
	GTR // >
	NEQ // <>
	LEQ // <=
	GEQ // >=

	ASSIGN    // :=
	COLON     // :
	COMMA     // ,
	PERIOD    // .
	ELLIPSIS  // ..
	SEMICOLON // ;

	LPAREN // (
	LBRACK // [

	RPAREN // )
	RBRACK // ]
	operator_end

	keyword_beg
	// Keywords, excluding STRING
	AND
	ARRAY
	AS
	ASM
	BEGIN
	CASE
	CLASS
	CONST
	CONSTRUCTOR
	DESTRUCTOR
	DISPINTERFACE
	DIV
	DO
	DOWNTO
	ELSE
	END
	EXCEPT
	EXPORTS
	FILE
	FINALIZATION
	FINALLY
	FOR
	FUNCTION
	GOTO
	IF
	IMPLEMENTATION
	IN
	INHERITED
	INITIALIZATION
	INLINE
	INTERFACE
	IS
	LABEL
	LIBRARY
	MOD
	NIL
	NOT
	OBJECT
	OF
	OR
	PACKED
	PROCEDURE
	PROGRAM
	PROPERTY
	RAISE
	RECORD
	REPEAT
	RESOURCESTRING
	SET
	SHL
	SHR
	THEN
	THREADVAR
	TO
	TRY
	TYPE
	UNIT
	UNTIL
	USES
	VAR
	WHILE
	WITH
	XOR

	// Directives
	ABSOLUTE
	ABSTRACT
	ASSEMBLER
	AUTOMATED
	CDECL
	CONTAINS7
	DEFAULT
	DELAYED11
	DEPRECATED
	DISPID
	DYNAMIC
	EXPERIMENTAL
	EXPORT
	EXTERNAL
	FAR1
	FINAL
	FORWARD
	HELPER8
	IMPLEMENTS
	INDEX
	INLINE2
	LIBRARY3
	LOCAL4
	MESSAGE
	NAME
	NEAR1
	NODEFAULT
	OPERATOR10
	OUT
	OVERLOAD
	OVERRIDE7
	PACKAGE
	PASCAL
	PLATFORM
	PRIVATE
	PROTECTED
	PUBLIC
	PUBLISHED
	READ
	READONLY
	REFERENCE9
	REGISTER
	REINTRODUCE
	REQUIRES7
	RESIDENT1
	SAFECALL
	SEALED5
	STATIC
	STDCALL
	STORED
	STRICT
	UNSAFE
	VARARGS
	VIRTUAL
	WINAPI6
	WRITE
	WRITEONLY
	keyword_end
)

var tokens = [...]string{
	ILLEGAL:    "ILLEGAL",
	EOF:        "EOF",
	COMMENT:    "COMMENT",
	CDIRECTIVE: "CDIRECTIVE",

	IDENT:   "IDENT",
	INTEGER: "INTEGER",
	FLOAT:   "FLOAT",
	CHAR:    "CHAR",
	STRING:  "STRING",

	AT:  "@",
	HAT: "^",

	ADD:  "+",
	SUB:  "-",
	MUL:  "*",
	FDIV: "/",

	EQL: "=",
	LSS: "<",
	GTR: ">",
	NEQ: "<>",
	LEQ: "<=",
	GEQ: ">=",

	ASSIGN:    ":=",
	COLON:     ":",
	COMMA:     ",",
	PERIOD:    ".",
	ELLIPSIS:  "..",
	SEMICOLON: ";",

	LPAREN: "(",
	LBRACK: "[",

	RPAREN: ")",
	RBRACK: "]",

	// Keywords
	AND:            "and",
	ARRAY:          "array",
	AS:             "as",
	ASM:            "asm",
	BEGIN:          "begin",
	CASE:           "case",
	CLASS:          "class",
	CONST:          "const",
	CONSTRUCTOR:    "constructor",
	DESTRUCTOR:     "destructor",
	DISPINTERFACE:  "dispinterface",
	DIV:            "div",
	DO:             "do",
	DOWNTO:         "downto",
	ELSE:           "else",
	END:            "end",
	EXCEPT:         "except",
	EXPORTS:        "exports",
	FILE:           "file",
	FINALIZATION:   "finalization",
	FINALLY:        "finally",
	FOR:            "for",
	FUNCTION:       "function",
	GOTO:           "goto",
	IF:             "if",
	IMPLEMENTATION: "implementation",
	IN:             "in",
	INHERITED:      "inherited",
	INITIALIZATION: "initialization",
	INLINE:         "inline",
	INTERFACE:      "interface",
	IS:             "is",
	LABEL:          "label",
	LIBRARY:        "library",
	MOD:            "mod",
	NIL:            "nil",
	NOT:            "not",
	OBJECT:         "object",
	OF:             "of",
	OR:             "or",
	PACKED:         "packed",
	PROCEDURE:      "procedure",
	PROGRAM:        "program",
	PROPERTY:       "property",
	RAISE:          "raise",
	RECORD:         "record",
	REPEAT:         "repeat",
	RESOURCESTRING: "resourcestring",
	SET:            "set",
	SHL:            "shl",
	SHR:            "shr",
	THEN:           "then",
	THREADVAR:      "threadvar",
	TO:             "to",
	TRY:            "try",
	TYPE:           "type",
	UNIT:           "unit",
	UNTIL:          "until",
	USES:           "uses",
	VAR:            "var",
	WHILE:          "while",
	WITH:           "with",
	XOR:            "xor",

	// Directives
	ABSOLUTE:     "absolute",
	ABSTRACT:     "abstract",
	ASSEMBLER:    "assembler",
	AUTOMATED:    "automated",
	CDECL:        "cdecl",
	CONTAINS7:    "contains7",
	DEFAULT:      "default",
	DELAYED11:    "delayed11",
	DEPRECATED:   "deprecated",
	DISPID:       "dispid",
	DYNAMIC:      "dynamic",
	EXPERIMENTAL: "experimental",
	EXPORT:       "export",
	EXTERNAL:     "external",
	FAR1:         "far1",
	FINAL:        "final",
	FORWARD:      "forward",
	HELPER8:      "helper8",
	IMPLEMENTS:   "implements",
	INDEX:        "index",
	INLINE2:      "inline2",
	LIBRARY3:     "library3",
	LOCAL4:       "local4",
	MESSAGE:      "message",
	NAME:         "name",
	NEAR1:        "near1",
	NODEFAULT:    "nodefault",
	OPERATOR10:   "operator10",
	OUT:          "out",
	OVERLOAD:     "overload",
	OVERRIDE7:    "override7",
	PACKAGE:      "package",
	PASCAL:       "pascal",
	PLATFORM:     "platform",
	PRIVATE:      "private",
	PROTECTED:    "protected",
	PUBLIC:       "public",
	PUBLISHED:    "published",
	READ:         "read",
	READONLY:     "readonly",
	REFERENCE9:   "reference9",
	REGISTER:     "register",
	REINTRODUCE:  "reintroduce",
	REQUIRES7:    "requires7",
	RESIDENT1:    "resident1",
	SAFECALL:     "safecall",
	SEALED5:      "sealed5",
	STATIC:       "static",
	STDCALL:      "stdcall",
	STORED:       "stored",
	STRICT:       "strict",
	UNSAFE:       "unsafe",
	VARARGS:      "varargs",
	VIRTUAL:      "virtual",
	WINAPI6:      "winapi6",
	WRITE:        "write",
	WRITEONLY:    "writeonly",
}

// String returns the string corresponding to the token tok.
// For operators, delimiters, and keywords the string is the actual
// token character sequence (e.g., for the token ADD, the string is
// "+"). For all other tokens the string corresponds to the token
// constant name (e.g. for the token IDENT, the string is "IDENT").
//
func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

// A set of constants for precedence-based expression parsing.
// Non-operators have lowest precedence, followed by operators
// starting with precedence 1 up to unary operators. The highest
// precedence serves as "catch-all" precedence for selector,
// indexing, and other operator and delimiter tokens.
//
const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 4
	HighestPrec = 5
)

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
//
func (op Token) Precedence() int {
	switch op {
	case EQL, NEQ, LSS, GTR, LEQ, GEQ, IN, IS:
		return 1
	case ADD, SUB, OR, XOR:
		return 2
	case MUL, FDIV, DIV, MOD, AND, SHL, SHR, AS:
		return 3
	}
	return LowestPrec
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

// Lookup maps an identifier to its keyword token or IDENT (if not a keyword).
//
func Lookup(ident string) Token {
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
//
func (tok Token) IsLiteral() bool { return literal_beg < tok && tok < literal_end }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
//
func (tok Token) IsOperator() bool {
	if operator_beg < tok && tok < operator_end {
		return true
	}
	switch tok {
	case EQL, NEQ, LSS, GTR, LEQ, GEQ, IN, IS, ADD, SUB, OR, XOR,
		MUL, FDIV, DIV, MOD, AND, SHL, SHR, AS, AT, HAT:
		return true
	}
	return false
}

// IsKeyword returns true for tokens corresponding to keywords;
// it returns false otherwise.
//
func (tok Token) IsKeyword() bool { return keyword_beg < tok && tok < keyword_end }
