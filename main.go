package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bradfitz/slice"
	"github.com/egonelbre/async"
)

var (
	nprocs = flag.Int("procs", 8, "number of parallel parsers to use")
)

type Reader struct {
	CompilerDefs []string
}

func cname(s string) string {
	return strings.ToLower(s)
}

type Ref struct {
	File  string
	Name  string
	CName string
}

var (
	rxProc = regexp.MustCompile(`(?i)\bprocedure\s+([a-z0-9\._]+)\b`)
	rxFunc = regexp.MustCompile(`(?i)\bfunction\s+([a-z0-9\._]+)\b`)
)

type Report struct {
	Funcs map[string][]Ref

	rxProc *regexp.Regexp
	rxFunc *regexp.Regexp
}

func NewReport() *Report {
	return &Report{
		Funcs:  make(map[string][]Ref, 1<<20),
		rxProc: rxProc.Copy(),
		rxFunc: rxFunc.Copy(),
	}
}

func (report *Report) AddFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	added := make(map[string]struct{})
	for _, match := range report.rxProc.FindAllSubmatch(data, -1) {
		name := string(match[1])
		cn := cname(name)

		if _, ok := added[cn]; ok {
			continue
		}
		added[cn] = struct{}{}

		report.Funcs[cn] = append(report.Funcs[cn], Ref{
			File:  filename,
			Name:  name,
			CName: cn,
		})
	}

	for _, match := range report.rxFunc.FindAllSubmatch(data, -1) {
		name := string(match[1])
		cn := cname(name)

		if _, ok := added[cn]; ok {
			continue
		}
		added[cn] = struct{}{}

		report.Funcs[cn] = append(report.Funcs[cn], Ref{
			File:  filename,
			Name:  name,
			CName: cn,
		})
	}

	return nil
}

func (a *Report) Combine(b *Report) {
	for name, refs := range b.Funcs {
		a.Funcs[name] = append(a.Funcs[name], refs...)
	}
}

func main() {
	flag.Parse()

	rootdir := flag.Arg(0)
	if rootdir == "" {
		fmt.Println("no directory specified")
		flag.PrintDefaults()
		os.Exit(1)
	}

	filenames := make(chan string, *nprocs)
	reports := make(chan *Report, *nprocs)
	errors := make(chan error, *nprocs)

	async.Spawn(*nprocs, func(id int) {
		report := NewReport()
		for filename := range filenames {
			err := report.AddFile(filename)
			if err != nil {
				errors <- err
			}
		}
		reports <- report
	}, func() { close(reports); close(errors) })

	go func() {
		err := filepath.Walk(rootdir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.ToLower(filepath.Ext(path)) != ".pas" {
				return nil
			}
			if strings.HasPrefix(info.Name(), ".") || strings.HasPrefix(info.Name(), "~") || strings.HasSuffix(info.Name(), "~") {
				return filepath.SkipDir
			}
			filenames <- path
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
		close(filenames)
	}()

	for err := range errors {
		fmt.Println(err)
	}

	total := NewReport()
	for sub := range reports {
		total.Combine(sub)
	}

	var exceeding [][]Ref
	for _, refs := range total.Funcs {
		if len(refs) >= 2 {
			exceeding = append(exceeding, refs)
		}
	}

	//sort.Sort(ByName(exceeding))
	slice.Sort(exceeding, func(i, j int) bool {
		return exceeding[i][0].CName < exceeding[j][0].CName
	})

	for _, refs := range exceeding {
		fmt.Println()
		fmt.Println(refs[0].Name)
		for _, ref := range refs {
			fmt.Println("\t", ref.File)
		}
	}
}

type ByName [][]Ref

func (refs ByName) Len() int           { return len(refs) }
func (refs ByName) Swap(i, j int)      { refs[i], refs[j] = refs[j], refs[i] }
func (refs ByName) Less(i, j int) bool { return refs[i][0].CName < refs[j][0].CName }
