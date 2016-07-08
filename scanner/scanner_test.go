package scanner_test

import (
	"testing"

	"github.com/raintreeinc/delphi/scanner"
	T "github.com/raintreeinc/delphi/token"
)

type elt struct {
	src  string
	toks []T.Token
}

var tests = [...]elt{
	{`p^+^p#13`, []T.Token{
		T.IDENT, T.HAT, T.ADD, T.CHAR, T.CHAR}},
	{`^TR`, []T.Token{
		T.HAT, T.IDENT}},
	{`type PPInteger = ^PInteger;`, []T.Token{
		T.TYPE, T.IDENT, T.EQL, T.HAT, T.IDENT, T.SEMICOLON}},
}

func TestScanner_Scan(t *testing.T) {
	for _, test := range tests {
		src := []byte(test.src)

		// Initialize the scanner.
		var s scanner.Scanner
		fset := T.NewFileSet()                          // positions are relative to fset
		file := fset.AddFile("", fset.Base(), len(src)) // register input "file"
		s.Init(file, src, nil /* no error handler */, scanner.ScanComments)

		for k, exp := range test.toks {
			_, tok, _ := s.Scan()
			if tok != exp {
				t.Errorf("%v: at %v expected %v got %v", test.src, k, exp, tok)
			}
		}
		_, tok, _ := s.Scan()
		if tok != T.EOF {
			t.Errorf("%v: at %v expected %v got %v", test.src, len(test.toks), T.EOF, tok)
		}
	}
}
