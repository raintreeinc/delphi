package test

import "github.com/raintreeinc/delphi/internal/cli"

const ShortDesc = "test units"

func Help(args []string) {
	cli.Helpf("Usage:\n")
	cli.Helpf("\t%s [arguments]\n\n", args[0])
}

func Main(args []string) {
	if len(args) <= 1 {
		Help(args)
		return
	}
}
