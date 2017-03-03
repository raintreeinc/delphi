package tokenize

import (
	"flag"

	"github.com/raintreeinc/delphi/internal/cli"
)

const ShortDesc = "tokenize file"

func Help(args []string) {
	cli.Helpf("Usage:\n")
	cli.Helpf("\t%s [filename]\n\n", args[0])
	cli.Helpf(`Arguments:
  -comments show comments
`)
}

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
	flags.Set.BoolVar(&flags.Help, "h", false, "show help")
	flags.Set.Parse(args[1:])
	flags.Files = flags.Set.Args()
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
				cli.Printf("\n")
			}
			cli.Printf("Tokenizing %q\n", file)
		}
		var state State
		state.Comments = flags.Comments
		state.Input = file
		state.Run()
	}
}
