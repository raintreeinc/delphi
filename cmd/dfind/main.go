// dfind is for matching, replacing and analysing based on regexps

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/egonelbre/async"
	"github.com/raintreeinc/delphi/internal/walk"
)

var (
	verbose     = flag.Bool("v", false, "verbose")
	batchfile   = flag.String("batch", "", "file describing all renames")
	nprocs      = flag.Int("procs", 8, "number of parallel parsers to use")
	write       = flag.Bool("w", false, "write changes to files")
	ignoreCase  = flag.Bool("i", false, "ignore case")
	ignoreSpace = flag.Bool("is", false, "ignore spaces when counting")

	// different modes
	count   = flag.Bool("count", false, "count number of occurrances")
	replace = flag.String("replace", "", "replace")
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) <= 0 {
		fmt.Println("PATTERN not specified")
		fmt.Println("Usage:")
		fmt.Println("	dfind [OPTIONS] PATTERN [PATH]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	match := args[0]
	globs := args[1:]
	if len(globs) == 0 {
		globs = []string{"."}
	}

	if *ignoreCase {
		match = "(?i)" + match
	}

	re, err := regexp.Compile(match)
	if err != nil {
		fmt.Println("Failed to compile PATTERN: %v", err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	// ensure that we do something
	if *replace == "" {
		*count = true
	}

	filenames := make(chan string, *nprocs)
	errors := make(chan error, *nprocs)
	counters := make(chan *Counter, *nprocs)
	go func() {
		walk.Globs(globs, filenames, errors)
		close(filenames)
	}()

	async.Spawn(*nprocs, func(id int) {
		total := NewCounter()

		r := re.Copy()
		for filename := range filenames {
			file, err := LoadFile(filename)
			if err != nil {
				errors <- fmt.Errorf("loading %v: %v", filename, err)
				continue
			}

			if *count {
				file.CountRegular(r, total)
			}
			if *replace != "" {
				file.Replace(r, *replace)
				if err := file.SaveChanges(); err != nil {
					errors <- err
				}
			}
		}

		counters <- total
	}, func() { close(errors); close(counters) })

	for err := range errors {
		fmt.Println(err)
	}

	Total := NewCounter()
	for counter := range counters {
		Total.Merge(counter)
	}

	if *count {
		var names []string
		for name := range Total.Actual {
			names = append(names, name)
		}
		sort.Strings(names)

		tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
		for _, name := range names {
			match := Total.Matches[name]
			fmt.Fprintf(tw, "'%v'\t%v\n", Total.Actual[name], match.Count)

			if *verbose {
				var files []string
				for file := range match.Files {
					files = append(files, file)
				}
				sort.Strings(files)

				for _, file := range files {
					fmt.Fprintf(tw, "    - %v\t%v\n", file, match.Files[file])
				}
				fmt.Fprintln(tw)
			}
		}
		tw.Flush()
	}

	if *verbose {
		fmt.Println("<DONE>")
	}
}

func (file *File) CountRegular(re *regexp.Regexp, counter *Counter) {
	re.ReplaceAllFunc(file.Source, func(match []byte) []byte {
		counter.Add(file.Path, string(match))
		return match
	})
}

func (file *File) Replace(re *regexp.Regexp, replacement string) {
	file.Modified = re.ReplaceAll(file.Source, []byte(replacement))
}

/* tabulation and counting */

type Counter struct {
	Total   int
	Matches map[string]*Match
	Actual  map[string]string
}

type Match struct {
	Count int
	Files map[string]int
}

func NewCounter() *Counter {
	return &Counter{
		Total:   0,
		Matches: make(map[string]*Match),
		Actual:  make(map[string]string),
	}
}

func NewMatch() *Match {
	return &Match{0, make(map[string]int)}
}

func (counter *Counter) Add(file string, match string) {
	canon := match
	if *ignoreCase {
		canon = strings.ToLower(canon)
	}
	if *ignoreSpace {
		canon = strings.Replace(canon, " ", "", -1)
		canon = strings.Replace(canon, "\t", "", -1)
		canon = strings.Replace(canon, "\n", "", -1)
	}

	if _, ok := counter.Actual[canon]; !ok {
		counter.Actual[canon] = match
		counter.Matches[canon] = NewMatch()
	}

	counter.Total++
	counter.Matches[canon].Add(file)
}

func (m *Match) Add(file string) {
	m.Count++
	m.Files[file]++
}

func (counter *Counter) Merge(other *Counter) {
	counter.Total += other.Total
	for canon, actual := range other.Actual {
		if _, ok := counter.Actual[canon]; !ok {
			counter.Actual[canon] = actual
			counter.Matches[canon] = NewMatch()
		}
	}

	for canon, match := range other.Matches {
		counter.Matches[canon].Merge(match)
	}
}

func (m *Match) Merge(other *Match) {
	m.Count += other.Count
	for file, count := range other.Files {
		m.Files[file] += count
	}
}

/* basic in memory file handling */

type File struct {
	Path     string
	Source   []byte
	Modified []byte
}

func LoadFile(path string) (*File, error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &File{path, src, src}, nil
}

func (file *File) Changed() bool { return !bytes.Equal(file.Source, file.Modified) }
func (file *File) SaveChanges() error {
	if !file.Changed() {
		return nil
	}
	if *write {
		err := ioutil.WriteFile(file.Path, file.Modified, 0755)
		if err == nil {
			fmt.Println("MODIFIED ", file.Path)
		}
		return err
	} else if *verbose {
		fmt.Println("<<<", file.Path, ">>>")
		fmt.Println(string(file.Modified))
	} else {
		fmt.Println("MODIFIES ", file.Path)
	}
	return nil
}
