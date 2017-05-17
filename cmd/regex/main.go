package regex

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"text/tabwriter"

	"github.com/egonelbre/async"
	"github.com/raintreeinc/delphi/delphi"
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

  -match   extract specific match
  -case    case sensitive
  -nospace ignore spaces when counting

  -w    write replacements to file
  -r    replacement for pattern

  -care check only files that match these globs
`)
}

type Flags struct {
	Count bool

	Plain   bool
	Verbose bool
	Help    bool

	Procs int
	Write bool

	Match         int
	CaseSensitive bool
	IgnoreSpace   bool

	Care walk.GlobsFlag

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

	flags.Set.BoolVar(&flags.Plain, "plain", false, "do not output count")

	flags.Set.IntVar(&flags.Procs, "n", 8, "number of workers")

	flags.Set.IntVar(&flags.Match, "match", 0, "extract specific match")
	flags.Set.BoolVar(&flags.CaseSensitive, "case", false, "case sensitive")
	flags.Set.BoolVar(&flags.IgnoreSpace, "nospace", false, "ignore spaces when counting")

	flags.Set.BoolVar(&flags.Write, "w", false, "write replacements to file")
	flags.Set.StringVar(&flags.Replacement, "r", "", "replacement for pattern")

	flags.Set.Var(&flags.Care, "care", "check only files that match these globs")

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

	if !flags.CaseSensitive {
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

	care := walk.IsDelphiFile
	if !flags.Care.IsEmpty() {
		care = flags.Care.Matches
	}

	// walk files
	go func() {
		walk.Globs(flags.Paths, filenames, errors, care)
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
				file.CountRegular(r, total, flags.Match, !flags.CaseSensitive, flags.IgnoreSpace)
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

		if flags.Plain {
			for _, name := range names {
				fmt.Fprintf(os.Stdout, "%v\n", delphi.Quote(total.Actual[name]))
			}
		} else {
			tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
			for _, name := range names {
				match := total.Matches[name]
				fmt.Fprintf(tw, "%v\t%v\n", delphi.Quote(total.Actual[name]), match.Count)

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
}
