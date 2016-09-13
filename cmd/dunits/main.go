package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/raintreeinc/delphi/scanner"
	"github.com/raintreeinc/delphi/token"
)

var (
	outfile   = flag.String("out", "", "output file")
	directory = flag.String("dir", "", "folder to search for units")
	verbose   = flag.Bool("v", false, "verbose output")
)

func main() {
	flag.Parse()

	dprfile := flag.Arg(0)
	if dprfile == "" {
		log.Println("dpr not specified")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *directory == "" {
		*directory = filepath.Dir(dprfile)
	}

	index, err := NewIndex(*directory)
	if err != nil {
		log.Fatal(err)
	}

	index.Build(dprfile)

	if *outfile == "" {
		*outfile = TrimExt(filepath.Base(dprfile)) + ".tgf"
	}

	file, err := os.Create(*outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	wr := bufio.NewWriter(file)
	defer wr.Flush()

	ext := strings.ToLower(filepath.Ext(*outfile))
	if ext == ".tgf" {
		WriteTGF(index, wr)
	} else if ext == ".dot" {
		WriteDOT(index, wr)
	} else {
		log.Fatal("Unknown file extension " + ext)
	}
}

func WriteDOT(index *Index, out io.Writer) (n int, err error) {
	write := func(format string, args ...interface{}) bool {
		if err != nil {
			return false
		}
		var x int
		x, err = fmt.Fprintf(out, format, args...)
		n += x
		return err == nil
	}

	write("digraph G{\n")
	for _, uses := range index.Uses {
		for _, use := range uses.Interface {
			write("\t%v -> %v;\n", uses.Unit, index.NormalName(use))
		}
		for _, use := range uses.Implementation {
			write("\t%v -> %v [style=dashed;dir=both;weight=0];\n", uses.Unit, index.NormalName(use))
		}
	}
	write("}\n")

	return
}

func WriteTGF(index *Index, out io.Writer) (n int, err error) {
	write := func(format string, args ...interface{}) bool {
		if err != nil {
			return false
		}
		var x int
		x, err = fmt.Fprintf(out, format, args...)
		n += x
		return err == nil
	}

	ids := make(map[string]int, len(index.Uses))

	id := 0
	for cunitname, use := range index.Uses {
		id++
		ids[cunitname] = id

		write("%v %v\n", id, use.Unit)
	}

	write("#\n")

	for cunitname, uses := range index.Uses {
		for _, use := range uses.Interface {
			write("%v %v\n", ids[cunitname], ids[strings.ToLower(use)])
		}
		for _, use := range uses.Implementation {
			write("%v %v\n", ids[cunitname], ids[strings.ToLower(use)])
		}
	}

	return 0, nil
}

type Index struct {
	Dir  string
	Path map[string]string
	Uses map[string]*UnitUses
}

type UnitUses struct {
	Unit           string
	Interface      []string
	Implementation []string
}

func NewIndex(dir string) (*Index, error) {
	index := &Index{
		Dir:  dir,
		Path: make(map[string]string),
		Uses: make(map[string]*UnitUses),
	}
	return index, index.load(dir)
}

func (index *Index) load(dir string) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := filepath.Base(path)
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "~") || strings.HasSuffix(path, "~") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".pas" || ext == ".dpr" {
			index.include(path)
		}

		return nil
	})

	return err
}

func (index *Index) include(path string) {
	name := filepath.Base(path)
	unitname := strings.ToLower(TrimExt(name))

	if _, duplicate := index.Path[unitname]; duplicate {
		log.Println("duplicate entry:", name)
		return
	}
	index.Path[unitname] = path
}

func (index *Index) Build(dprfile string) {
	queue := []string{}
	queue = append(queue, TrimExt(filepath.Base(dprfile)))

	for len(queue) > 0 {
		var unit string
		unit, queue = queue[len(queue)-1], queue[:len(queue)-1]

		uses := index.Load(unit)
		if uses == nil {
			continue
		}

		for _, use := range uses.Interface {
			if !index.IsLoaded(use) {
				queue = append(queue, use)
			}
		}

		for _, use := range uses.Implementation {
			if !index.IsLoaded(use) {
				queue = append(queue, use)
			}
		}
	}
}

func (index *Index) IsLoaded(unitname string) bool {
	cunitname := strings.ToLower(unitname)
	_, loaded := index.Uses[cunitname]
	return loaded
}

func (index *Index) Load(unitname string) *UnitUses {
	if index.IsLoaded(unitname) {
		return nil
	}
	cunitname := strings.ToLower(unitname)

	uses := &UnitUses{}
	uses.Unit = unitname
	index.Uses[cunitname] = uses

	unitpath, ok := index.Path[cunitname]
	if !ok {
		log.Printf("Did not find path for %v\n", unitname)
		return nil
	}

	src, err := ioutil.ReadFile(unitpath)
	if err != nil {
		log.Printf("Failed to read %v: %v", unitname, err)
		return nil
	}

	// Initialize the scanner.
	var s scanner.Scanner
	fset := token.NewFileSet()                      // positions are relative to fset
	file := fset.AddFile("", fset.Base(), len(src)) // register input "file"

	var flags scanner.Mode

	s.Init(file, src, func(pos token.Position, msg string) {
		if *verbose {
			log.Println("%s\tERROR\t%s\n", pos, msg)
		}
	}, flags)

	state := 1
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		_ = pos

		if tok == token.IMPLEMENTATION {
			state = 2
		} else if tok == token.IDENT {
			_, isunit := index.Path[strings.ToLower(lit)]
			if !isunit {
				continue
			}

			if state == 1 {
				uses.Interface = append(uses.Interface, lit)
			} else if state == 2 {
				uses.Implementation = append(uses.Implementation, lit)
			}
		}
	}

	return uses
}

func (index *Index) NormalName(name string) string {
	use, ok := index.Uses[strings.ToLower(name)]
	if !ok {
		return ""
	}
	return use.Unit
}

func TrimExt(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}
