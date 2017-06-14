package main

import (
	"os"
	"strings"

	"github.com/raintreeinc/delphi/cmd/regex"
	"github.com/raintreeinc/delphi/cmd/test"
	"github.com/raintreeinc/delphi/cmd/tokenize"
	"github.com/raintreeinc/delphi/cmd/uses"
	"github.com/raintreeinc/delphi/internal/cli"
)

type Command struct {
	Name      string
	ShortDesc string
	Main      func(args []string)
	Help      func(args []string)
}

var Commands []Command

func FindCommand(name string) *Command {
	if name == "" {
		return nil
	}

	for _, cmd := range Commands {
		if cmd.Name == name {
			return &cmd
		}
	}
	return nil
}

func CommandHelp(args []string) {
	if len(args) <= 1 {
		HelpHelp(args)
		os.Exit(2)
	}

	cmdname := args[1]
	cmd := FindCommand(cmdname)
	if cmd == nil {
		cli.Errorf("No command named %q\n\n", cmdname)
		HelpHelp(args)
		os.Exit(2)
	}

	args[0] = strings.TrimSuffix(args[0], "help") + cmdname

	cmd.Help(args)
	os.Exit(2)
}

func PrintCommands() {
	cli.Helpf("Commands are:\n")
	for _, cmd := range Commands {
		cli.Helpf("    %-8s %s\n", cmd.Name, cmd.ShortDesc)
	}
	cli.Helpf("\n")
}

func HelpHelp(args []string) {
	cli.Helpf("Usage:\n")
	cli.Helpf("\t%s [command]\n\n", args[0])
	PrintCommands()
}

func Help(args []string) {
	cli.Helpf("Usage:\n")
	cli.Helpf("\t%s command [arguments]\n\n", args[0])
	PrintCommands()
	cli.Helpf("Use \"%s help [command]\" for more information about a command.\n", args[0])
}

func main() {
	Commands = []Command{
		{"test", test.ShortDesc, test.Main, test.Help},
		{"uses", uses.ShortDesc, uses.Main, uses.Help},
		{"regex", regex.ShortDesc, regex.Main, regex.Help},
		{},
		{"tokenize", tokenize.ShortDesc, tokenize.Main, tokenize.Help},
		{"help", "print help about a command", CommandHelp, HelpHelp},
	}

	if len(os.Args) <= 1 {
		Help(os.Args)
		os.Exit(2)
	}

	cmdname := os.Args[1]
	args := append([]string{os.Args[0] + " " + os.Args[1]}, os.Args[2:]...)

	cmd := FindCommand(cmdname)
	if cmd == nil {
		cli.Errorf("No command named %q\n\n", cmdname)
		Help(os.Args)
		os.Exit(2)
	}

	cmd.Main(args)
}
