package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/raintreeinc/delphi/scanner"
	"github.com/raintreeinc/delphi/token"

	"github.com/kr/pretty"
)

var (
	verbose = flag.Int("v", 0, "verbosity level")
)

func verbosity(v int) bool {
	return v <= *verbose
}

type Unit struct {
	Name           Ident
	Interface      Declarations
	Implementation Declarations
}

type Declarations struct {
	Pos     token.Pos
	Classes []Class
	Records []Record
	Procs   []Proc
}

type Class struct {
	Name  Ident
	Procs []Proc
}

type Record struct {
	Name Ident
}

type ProcFlag byte

const (
	Procedure ProcFlag = iota
	Function
	Reintroduce
	Overload
	Virtual
	Override
	Abstract
)

type Proc struct {
	Name  Ident
	Flags ProcFlag
}

type Ident struct {
	Pos  token.Pos
	Name string
}

func main() {
	flag.Parse()

	rootfile := flag.Arg(0)
	if rootfile == "" {
		fmt.Println("no file specified")
		flag.PrintDefaults()
		os.Exit(1)
	}

	src, err := ioutil.ReadFile(rootfile)
	if err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Initialize the scanner.
	var s scanner.Scanner
	fset := token.NewFileSet()                      // positions are relative to fset
	file := fset.AddFile("", fset.Base(), len(src)) // register input "file"
	s.Init(file, src, func(pos token.Position, msg string) {
		fmt.Printf("%s\tERROR\t%s\n", pos, msg)
	}, 0)

	parser := &UnitParser{}
	parser.FileSet = fset
	parser.Scanner = &s
	parser.Unit = &Unit{}
	parser.ParseUnit()

	pretty.Println(parser.Unit)
}

type UnitParser struct {
	FileSet *token.FileSet
	Scanner *scanner.Scanner
	Unit    *Unit
	Current Token
	Past    [2]Token
}

type Token struct {
	Pos     token.Pos
	Token   token.Token
	Literal string
}

func (p *UnitParser) Next() Token {
	pos, tok, lit := p.Scanner.Scan()
	if verbosity(5) {
		fmt.Printf("%s\t%s\t%q\n", p.FileSet.Position(pos), tok, lit)
	}

	p.Past[1], p.Past[0] = p.Past[0], p.Current
	p.Current = Token{pos, tok, lit}
	return p.Current
}

func (p *UnitParser) FindNext(target token.Token) (Token, bool) {
	// Repeated calls to Scan yield the token sequence found in the input.
	for {
		t := p.Next()
		if t.Token == target {
			return t, true
		}
		if t.Token == token.EOF {
			return t, false
		}
	}
	panic("unreachable")
}

func (p *UnitParser) FindAny(targets ...token.Token) (Token, bool) {
	// Repeated calls to Scan yield the token sequence found in the input.
	for {
		t := p.Next()
		for _, target := range targets {
			if t.Token == target {
				return t, true
			}
		}
		if t.Token == token.EOF {
			panic("expected something better")
			return t, false
		}
	}
	panic("unreachable")
}

func (p *UnitParser) NextIdent() Ident {
	t := p.Next()
	if t.Token != token.IDENT {
		return Ident{p.Past[0].Pos, ""}
	}
	return Ident{t.Pos, t.Literal}
}

func (p *UnitParser) ParseUnit() {
	var t Token

	p.FindNext(token.UNIT)
	p.Unit.Name = p.NextIdent()

	t, _ = p.FindNext(token.INTERFACE)
	p.Unit.Interface.Pos = t.Pos

	for {
		t, _ = p.FindAny(
			token.CLASS, token.INTERFACE, token.RECORD,
			token.PROCEDURE, token.FUNCTION,

			token.IMPLEMENTATION,
		)
		if t.Token == token.IMPLEMENTATION || t.Token == token.EOF {
			break
		}

		if verbosity(5) {
			fmt.Println("==========================")
		}

		switch t.Token {
		case token.CLASS:
			class := Class{}
			if p.Past[1].Token == token.IDENT && p.Past[0].Token == token.EQL {
				class.Name = Ident{p.Past[1].Pos, p.Past[1].Literal}
			}
			p.Unit.Interface.Classes = append(p.Unit.Interface.Classes, class)
			if t = p.Next(); t.Token != token.SEMICOLON {
				p.FindNext(token.END)
			}
			fmt.Println(t)
		case token.INTERFACE:
			class := Class{}
			if p.Past[1].Token == token.IDENT && p.Past[0].Token == token.EQL {
				class.Name = Ident{p.Past[1].Pos, p.Past[1].Literal}
			}
			p.Unit.Interface.Classes = append(p.Unit.Interface.Classes, class)
			if t = p.Next(); t.Token != token.SEMICOLON {
				p.FindNext(token.END)
			}
		case token.RECORD:
			rec := Record{}
			if p.Past[1].Token == token.IDENT && p.Past[0].Token == token.EQL {
				rec.Name = Ident{p.Past[1].Pos, p.Past[1].Literal}
			} else {
				rec.Name = Ident{t.Pos, ""}
			}
			p.Unit.Interface.Records = append(p.Unit.Interface.Records, rec)
			p.FindNext(token.END)

		case token.PROCEDURE, token.FUNCTION:
			proc := Proc{}
			proc.Name = p.NextIdent()
			if t.Token == token.FUNCTION {
				proc.Flags = Function
			} else {
				proc.Flags = Procedure
			}
			p.Unit.Interface.Procs = append(p.Unit.Interface.Procs, proc)
		}
	}
	p.Unit.Implementation.Pos = t.Pos
}
