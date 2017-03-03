package tokenize

import (
	"flag"
	"fmt"
)

const ShortDesc = "tokenize file"

func Flags(args []string) *flag.FlagSet {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.Parse(args[1:])
	return flags
}

func Help(args []string) {
	fmt.Println("Usage:")
	fmt.Printf("    %s [filename]\n", args[0])
}

func Main(args []string) {
	if len(args) <= 1 {
		Help(args)
		return
	}

}
