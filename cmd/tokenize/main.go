package tokenize

import (
	"flag"
	"fmt"
)

const ShortDesc = "tokenize file"

var comments bool

type Flags struct {
	Comments bool
	Help     bool

	Set *flag.FlagSet
}

func ParseFlags(args []string) *Flags {
	flags := &Flags{}
	flags.Set = flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.Set.BoolVar(&flags.Comments, "comments", false, "include comments")
	flags.Set.BoolVar(&flags.Help, "help", false, "show help")
	flags.Set.Parse(args[1:])
	return flags
}

func Help(args []string) {
	fmt.Println("Usage:")
	fmt.Printf("\t%s [filename]\n", args[0])
	fmt.Println()
	fmt.Println("Arguments:")
	ParseFlags(args).Set.PrintDefaults()
}

func Main(args []string) {
	if len(args) <= 1 {
		Help(args)
		return
	}

	flags := ParseFlags(args)
	if flags.Help {
		Help(args)
		return
	}

}
