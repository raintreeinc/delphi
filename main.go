package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/raintreeinc/delphi/cmd/test"
	"github.com/raintreeinc/delphi/cmd/tokenize"
)

type Command struct {
	Name      string
	ShortDesc string
	Main      func(args []string)
	Help      func(args []string)
}

var Commands []Command

func FindCommand(name string) *Command {
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
		fmt.Fprintf(os.Stderr, "No command named %q\n", cmdname)
		fmt.Fprintln(os.Stderr, "")
		HelpHelp(args)
		os.Exit(2)
	}

	args[0] = strings.TrimSuffix(args[0], "help") + cmdname

	cmd.Help(args)
	os.Exit(2)
}

func HelpHelp(args []string) {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintf(os.Stderr, "    %s [command]\n", args[0])
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands are:")
	for _, cmd := range Commands {
		fmt.Printf("    %-8s %s\n", cmd.Name, cmd.ShortDesc)
	}
}

func Help(args []string) {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintf(os.Stderr, "    %s command [arguments]\n", args[0])
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands are:")
	for _, cmd := range Commands {
		fmt.Printf("    %-8s %s\n", cmd.Name, cmd.ShortDesc)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Use \"%s help [command]\" for more information about a command.\n", args[0])
}

func main() {
	Commands = []Command{
		{"test", test.ShortDesc, test.Main, test.Help},
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
		fmt.Fprintf(os.Stderr, "No command named %q\n", cmdname)
		fmt.Fprintln(os.Stderr, "")
		Help(os.Args)
		os.Exit(2)
	}

	cmd.Main(args)
}
