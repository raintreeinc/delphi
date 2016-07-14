// dtokenize prints tokens inside a pas file

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/raintreeinc/delphi/scanner"
	"github.com/raintreeinc/delphi/token"
)

var (
	comments = flag.Bool("comments", false, "include comments")
)

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

	var flags scanner.Mode
	if *comments {
		flags = scanner.ScanComments
	}

	s.Init(file, src, func(pos token.Position, msg string) {
		fmt.Printf("%s\tERROR\t%s\n", pos, msg)
	}, flags)

	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		fmt.Printf("%s\t%s\t%q\n", fset.Position(pos), tok, lit)
	}
}
