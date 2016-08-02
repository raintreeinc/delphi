// drename is inteded for batch renaming Delphi variables.
//
// This utility is useful for cleaning up or restructuring old and large code-bases.
// Especially where you have conditional compiling.
//
// Warning: drename currently is not context-sensitive hence it may rename
// more than you wanted. And mostly is hacked together to serve a particular
// need. Eventually it will be adjusted to be context sensitive.

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/egonelbre/async"
	"github.com/raintreeinc/delphi/internal/walk"
	"github.com/raintreeinc/delphi/scanner"
	"github.com/raintreeinc/delphi/token"
)

var (
	verbose   = flag.Int("v", 0, "verbosity level")
	batchfile = flag.String("batch", "", "file describing all renames")
	nprocs    = flag.Int("procs", 8, "number of parallel parsers to use")
	write     = flag.Bool("w", false, "write changes to files")
)

func main() {
	flag.Parse()

	globs := flag.Args()
	if len(globs) == 0 {
		fmt.Println("Error: globs not specified.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	batch, err := LoadBatchFile(*batchfile)
	if err != nil {
		fmt.Printf("Error loading mapping: %s\n", err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	errs := batch.CheckDuplicates()
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err)
		}
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
			err := process(batch, filename)
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

func process(batch *BatchRename, filename string) error {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Initialize the scanner.
	var sc scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile(filename, fset.Base(), len(src))
	sc.Init(file, src, func(pos token.Position, msg string) { fmt.Printf("%s\tERROR\t%s\n", pos, msg) }, 0)

	var out bytes.Buffer
	{ // main processing
		uses := []string{}
		start := 0

		for {
			pos, tok, lit := sc.Scan()

			// emit unprocessed content
			if offset := fset.Position(pos).Offset; start < offset {
				out.Write(src[start:offset])
				start = offset
			}
			if tok == token.EOF {
				break
			} else if tok != token.IDENT {
				continue
			}

			lit = Canonical(lit)

			// process uses list
			if batch.IsUnit(lit) {
				contains := false
				for _, unit := range uses {
					if unit == lit {
						contains = true
						break
					}
				}
				if !contains {
					uses = append(uses, Canonical(lit))
				}
				continue
			}

			// process replacements
			if repl, ok := batch.Lookup(uses, lit); ok {
				start += len(lit)
				out.WriteString(repl)
				continue
			}
		}
	}

	if !bytes.Equal(src, out.Bytes()) {
		if *write {
			fmt.Println("MODIFIED " + filename)
			ioutil.WriteFile(filename, out.Bytes(), 0755)
		} else if *verbose > 1 {
			fmt.Println("<--- " + filename + " --->\n" + out.String())
		} else {
			fmt.Println("MODIFIES " + filename)
		}
	}
	return nil
}

type BatchRename struct {
	Unit map[string]Mapping
}
type Mapping map[string]string

func LoadBatchFile(file string) (*BatchRename, error) {
	m := &BatchRename{}
	_, err := toml.DecodeFile(file, m)
	if err != nil {
		return nil, err
	}

	m.canonicalize()
	return m, nil
}

func (batch *BatchRename) CheckDuplicates() []error {
	dups := []error{}
	duplicate := map[string]bool{}
	for unit, mapping := range batch.Unit {
		for _, name := range mapping {
			cname := strings.ToLower(name)
			if duplicate[cname] {
				dups = append(dups, errors.New(fmt.Sprintf("%s: %s", unit, name)))
			}
			duplicate[cname] = true
		}
	}

	return dups
}

// Converts all source identifiers to Canonical form
// NB: calling this is necessary for IsUnit/Lookup to work properly
func (batch *BatchRename) canonicalize() {
	units := make(map[string]Mapping, len(batch.Unit))
	for unit, mapping := range batch.Unit {
		cmapping := make(Mapping, len(mapping))
		for ident, repl := range mapping {
			cmapping[Canonical(ident)] = repl
		}
		units[Canonical(unit)] = cmapping
	}
	batch.Unit = units
}

// Looks whether an identifier is potentially a unit
// NB: ident must be both in canonical form.
func (batch *BatchRename) IsUnit(ident string) bool {
	_, ok := batch.Unit[ident]
	return ok
}

// Looks whether an identifier must be replaced.
// NB: uses and ident must be both in canonical form.
func (batch *BatchRename) Lookup(uses []string, ident string) (replacement string, ok bool) {
	for _, unit := range uses {
		mapping, ok := batch.Unit[unit]
		if !ok {
			continue
		}
		if repl, ok := mapping[ident]; ok {
			return repl, ok
		}
	}
	return "", false
}

func Canonical(name string) string { return strings.ToLower(name) }
