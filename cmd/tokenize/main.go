package tokenize

import (
	"flag"
	"fmt"
	"os"
)

const ShortDesc = "tokenize file"

var comments bool

type Flags struct {
	Comments bool
	Help     bool
	Files    []string

	Set *flag.FlagSet
}

func (flags *Flags) Parse(args []string) {
	flags.Set = flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.Set.BoolVar(&flags.Comments, "comments", false, "include comments")
	flags.Set.BoolVar(&flags.Help, "help", false, "show help")
	flags.Set.Parse(args[1:])
	flags.Files = flags.Set.Args()
}

func Help(args []string) {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintf(os.Stderr, "\t%s [filename]\n", args[0])
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Arguments:")

	var flags Flags
	flags.Parse(args)
	flags.Set.PrintDefaults()
}

func Main(args []string) {
	var flags Flags
	flags.Parse(args)
	if flags.Help || len(flags.Files) == 0 {
		Help(args)
		return
	}

	for i, file := range flags.Files {
		if len(flags.Files) > 1 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("Tokenizing %q\n", file)
		}
		var state State
		state.Comments = flags.Comments
		state.Input = file
		state.Run()
	}
}
