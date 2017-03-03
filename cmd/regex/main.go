package regex

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"text/tabwriter"

	"github.com/egonelbre/async"
	"github.com/raintreeinc/delphi/internal/cli"
	"github.com/raintreeinc/delphi/internal/walk"
)

const ShortDesc = "count or replace using regular expressions"

func Help(args []string) {
	cli.Helpf("Usage:\n")
	cli.Helpf("\t%s [arguments] pattern path...\n\n", args[0])
	cli.Helpf(`Arguments:
  -v    verbose output
  -n    number of workers (default 8)

  -i    ignore case
  -is   ignore spaces when counting

  -w    write replacements to file
  -r    replacement for pattern
`)
}

type Flags struct {
	Count bool

	Verbose bool
	Help    bool

	Procs int
	Write bool

	IgnoreCase  bool
	IgnoreSpace bool

	Pattern     string
	Replacement string
	Paths       []string

	Set *flag.FlagSet
}

func (flags *Flags) Parse(args []string) {
	flags.Set = flag.NewFlagSet(args[0], flag.ExitOnError)

	flags.Set.BoolVar(&flags.Help, "help", false, "show help")
	flags.Set.BoolVar(&flags.Help, "h", false, "show help")
	flags.Set.BoolVar(&flags.Verbose, "v", false, "verbose output")

	flags.Set.IntVar(&flags.Procs, "n", 8, "number of workers")

	flags.Set.BoolVar(&flags.IgnoreCase, "i", false, "ignore case")
	flags.Set.BoolVar(&flags.IgnoreSpace, "is", false, "ignore spaces when counting")

	flags.Set.BoolVar(&flags.Write, "w", false, "write replacements to file")
	flags.Set.StringVar(&flags.Replacement, "r", "", "replacement for pattern")

	flags.Set.Parse(args[1:])

	if args := flags.Set.Args(); len(args) >= 1 {
		flags.Pattern = args[0]
		flags.Paths = args[1:]
	}
}

func Main(args []string) {
	var flags Flags
	flags.Parse(args)
	if flags.Help || flags.Pattern == "" {
		Help(args)
		return
	}

	if len(flags.Paths) == 0 {
		flags.Paths = []string{"."}
	}
	if flags.Replacement == "" {
		flags.Count = true
	}

	if flags.IgnoreCase {
		flags.Pattern = "(?i)" + flags.Pattern
	}

	re, err := regexp.Compile(flags.Pattern)
	if err != nil {
		cli.Errorf("invalid pattern %q: %s\n", flags.Pattern, err)
		os.Exit(1)
	}

	filenames := make(chan string, flags.Procs)
	errors := make(chan error, flags.Procs)
	counters := make(chan *Counter, flags.Procs)

	// walk files
	go func() {
		walk.Globs(flags.Paths, filenames, errors)
		close(filenames)
	}()

	// count/replace
	async.Spawn(flags.Procs, func(id int) {
		total := NewCounter()
		r := re.Copy()
		for filename := range filenames {
			file, err := LoadFile(filename)
			if err != nil {
				errors <- err
				continue
			}

			if flags.Count {
				file.CountRegular(r, total, flags.IgnoreCase, flags.IgnoreSpace)
			}
			if flags.Replacement != "" {
				file.Replace(r, flags.Replacement)

				if file.Changed() {
					if flags.Write {
						if err := file.WriteChanges(); err != nil {
							errors <- err
						}
						cli.Printf("modified %q\n", file.Path)
					} else {
						cli.Printf("modifies %q\n", file.Path)
					}
				}
			}
		}
		counters <- total
	}, func() { close(errors); close(counters) })

	for err := range errors {
		cli.Errorf("%s\n", err)
	}

	total := NewCounter()
	for counter := range counters {
		total.Merge(counter)
	}

	if flags.Count {
		var names []string
		for name := range total.Actual {
			names = append(names, name)
		}
		sort.Strings(names)

		tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
		for _, name := range names {
			match := total.Matches[name]
			fmt.Fprintf(tw, "'%v'\t%v\n", total.Actual[name], match.Count)

			if flags.Verbose {
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
}
