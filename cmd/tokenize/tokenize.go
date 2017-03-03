package tokenize

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/raintreeinc/delphi/scanner"
	"github.com/raintreeinc/delphi/token"
)

type State struct {
	Input    string
	Comments bool

	source  []byte
	files   *token.FileSet
	file    *token.File
	scanner scanner.Scanner
}

func (state *State) Run() {
	var err error
	state.source, err = ioutil.ReadFile(state.Input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed read input: %s", err)
		os.Exit(1)
	}

	state.files = token.NewFileSet()
	state.file = state.files.AddFile("", state.files.Base(), len(state.source))

	var flags scanner.Mode
	if state.Comments {
		flags = scanner.ScanComments
	}

	state.scanner.Init(state.file, state.source,
		func(pos token.Position, msg string) {
			fmt.Printf("ERROR: %s: %s", pos, msg)
		}, flags)

	for {
		pos, tok, lit := state.scanner.Scan()
		if tok == token.EOF {
			break
		}

		fmt.Printf("%s\t%s\t%q\n", state.files.Position(pos), tok, lit)
	}
}
