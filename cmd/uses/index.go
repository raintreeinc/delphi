package uses

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/raintreeinc/delphi/scanner"
	"github.com/raintreeinc/delphi/token"
)

type Index struct {
	Verbose       bool
	InterfaceOnly bool

	RootFiles []string

	Path    map[string]string
	IncPath map[string]string
	Uses    map[string]*UnitUses
}

type UnitUses struct {
	Unit           string
	Interface      []string // case insensitive sorted names
	Implementation []string // case insensitive sorted names
}

func NewIndex() *Index {
	return &Index{
		Path:    make(map[string]string),
		IncPath: make(map[string]string),
		Uses:    make(map[string]*UnitUses),
	}
}

func (index *Index) AddSourceDir(dir string) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		abs, _ := filepath.Abs(path)
		if abs != "" {
			path = abs
		}

		name := filepath.Base(path)
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "~") || strings.HasSuffix(path, "~") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".inc" {
			index.addIncludePath(path)
		}
		if ext == ".pas" || ext == ".dpr" {
			index.addSourcePath(path)
		}

		return nil
	})

	return err
}

func (index *Index) addSourcePath(path string) {
	name := filepath.Base(path)
	unitname := strings.ToLower(trimExt(name))

	if _, duplicate := index.Path[unitname]; duplicate {
		if index.Verbose {
			log.Println("duplicate entry:", name)
		}
		return
	}
	index.Path[unitname] = path
}

func (index *Index) addIncludePath(path string) {
	name := filepath.Base(path)
	unitname := strings.ToLower(trimExt(name))

	if _, duplicate := index.IncPath[unitname]; duplicate {
		if index.Verbose {
			log.Println("duplicate entry:", name)
		}
		return
	}
	index.IncPath[unitname] = path
}

func (index *Index) Build(rootfiles []string) {
	queue := []string{}
	for _, rootfile := range rootfiles {
		name := trimExt(filepath.Base(rootfile))
		queue = append(queue, name)
		index.RootFiles = append(index.RootFiles, name)
	}

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

	uses := &UnitUses{}
	uses.Unit = unitname
	index.Uses[strings.ToLower(unitname)] = uses

	unitpath, ok := index.Path[strings.ToLower(unitname)]
	if !ok {
		log.Printf("Did not find path for %v\n", unitname)
		return nil
	}

	// Initialize the scanner.
	index.scanUses(uses, unitpath, 1)

	return uses
}

func (index *Index) handleInclude(uses *UnitUses, directive string, state int) {
	p := strings.IndexRune(directive, ' ')
	name := strings.Trim(directive[p:], "{}'\" ")

	includepath, ok := index.IncPath[strings.ToLower(trimExt(name))]
	if !ok {
		if index.Verbose {
			log.Printf("Failed to include %v: %q", name, directive)
		}
		return
	}

	index.scanUses(uses, includepath, state)
}

func (index *Index) scanUses(uses *UnitUses, unitpath string, state int) {
	src, err := ioutil.ReadFile(unitpath)
	if err != nil {
		log.Printf("Failed to read %v: %v", unitpath, err)
		return
	}

	cunitname := strings.ToLower(uses.Unit)

	scanner.Scan(src, 0, func(tok token.Token, lit string) error {
		if tok == token.CDIRECTIVE {
			llit := strings.ToLower(lit)
			if strings.HasPrefix(llit, "{$i ") ||
				strings.HasPrefix(llit, "{$include ") {
				index.handleInclude(uses, lit, state)
			}
			return nil
		}

		if tok == token.IMPLEMENTATION {
			state = 2
		} else if tok == token.IDENT {
			cusename := strings.ToLower(lit)
			if cusename == cunitname {
				return nil
			}

			_, isunit := index.Path[cusename]
			if !isunit {
				return nil
			}

			if state == 1 {
				uses.Interface = includeString(uses.Interface, lit)
			} else if state == 2 {
				if !index.InterfaceOnly {
					uses.Implementation = includeString(uses.Implementation, lit)
				}
			}
		}
		return nil
	}, func(pos token.Position, msg string) {
		if index.Verbose {
			log.Printf("%s: %s\tERROR\t%s\n", unitpath, pos, msg)
		}
	})
}

func (index *Index) NormalName(name string) string {
	use, ok := index.Uses[strings.ToLower(name)]
	if !ok {
		return ""
	}
	return use.Unit
}
