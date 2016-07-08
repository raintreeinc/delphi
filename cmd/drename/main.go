package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/egonelbre/async"
	"github.com/raintreeinc/delphi/scanner"
	"github.com/raintreeinc/delphi/token"
	"github.com/raintreeinc/delphi/walk"
)

var (
	verbose = flag.Int("v", 0, "verbosity level")
	mapping = flag.String("mapping", "", "file describing all renames")
	unit    = flag.String("unit", "", "unit where the identifiers are defined")
	nprocs  = flag.Int("procs", 8, "number of parallel parsers to use")
	write   = flag.Bool("w", false, "write changes to files")
)

func main() {
	flag.Parse()

	globs := flag.Args()
	if len(globs) == 0 {
		fmt.Println("Error: globs not specified.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	mapping, err := LoadMapping(*mapping)
	if err != nil {
		fmt.Printf("Error loading mapping: %s\n", err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	filenames := make(chan string, *nprocs)
	errors := make(chan error)
	go func() {
		walk.Globs(globs, filenames, errors)
		close(filenames)
	}()

	async.Spawn(*nprocs, func(id int) {
		for filename := range filenames {
			err := process(mapping, filename)
			if err != nil {
				errors <- fmt.Errorf("%v: %v", filename, err)
			}
		}
	}, func() { close(errors) })

	for err := range errors {
		fmt.Println(err)
	}

	fmt.Println("<DONE>")
}

var rxTarget = regexp.MustCompile(`(?i)\brtScreenBuffer\b`)

func process(mapping *Mapping, filename string) error {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Initialize the scanner.
	var sc scanner.Scanner
	fset := token.NewFileSet()                            // positions are relative to fset
	file := fset.AddFile(filename, fset.Base(), len(src)) // register input "file"
	sc.Init(file, src, func(pos token.Position, msg string) { fmt.Printf("%s\tERROR\t%s\n", pos, msg) }, 0)

	renamer := &Renamer{
		Unit:    *unit,
		FileSet: fset,
		Scanner: &sc,
		Source:  src,
		Output:  bytes.Buffer{},
		Mapping: mapping,
	}
	renamer.Process()

	if !bytes.Equal(src, renamer.Output.Bytes()) {
		fmt.Println("CHANGING", filename)
		if *write {
			ioutil.WriteFile(filename, renamer.Output.Bytes(), 0755)
		}
	}
	return nil
}

type Mapping struct {
	Replace map[string]string
}

type Renamer struct {
	Unit    string
	FileSet *token.FileSet
	Scanner *scanner.Scanner
	Source  []byte
	Output  bytes.Buffer
	Mapping *Mapping
}

func (renamer *Renamer) Process() {
	found := renamer.Unit == ""
	unit := strings.ToLower(renamer.Unit)

	start := 0
	for {
		pos, tok, lit := renamer.Scanner.Scan()

		offset := renamer.FileSet.Position(pos).Offset
		if start < offset {
			renamer.Output.Write(renamer.Source[start:offset])
			start = offset
		}
		if tok == token.EOF {
			break
		}

		if !found && tok == token.IDENT {
			found = unit == strings.ToLower(lit)
		}
		if !found {
			continue
		}

		if tok == token.IDENT {
			if renamed, ok := renamer.Mapping.Replace[strings.ToLower(lit)]; ok {
				start += len(lit)
				renamer.Output.WriteString(renamed)
			}
		} else if tok == token.COMMENT {

		}
	}
}

func LoadMapping(file string) (*Mapping, error) {
	m := &Mapping{make(map[string]string, 1000)}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	rxSeparator := regexp.MustCompile("[\t ]+")
	text := string(data)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}

		toks := rxSeparator.Split(line, -1)
		if len(toks) == 0 {
			continue
		} else if len(toks) != 2 {
			return nil, fmt.Errorf("invalid number of replacements for %v", toks)
		}
		from, to := strings.TrimSpace(toks[0]), strings.TrimSpace(toks[1])
		m.Replace[strings.ToLower(from)] = to
	}
	return m, nil
}
