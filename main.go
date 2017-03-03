package main

import (
	"fmt"
	"os"

	"github.com/raintreeinc/delphi/cmd/test"
)

type Command struct {
	Name      string
	ShortDesc string
	Main      func(args []string)
	Help      func()
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
		HelpHelp()
		os.Exit(1)
	}

	cmdname := args[1]
	cmd := FindCommand(cmdname)
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "No command named %q\n", cmdname)
		fmt.Fprintln(os.Stderr, "")
		HelpHelp()
		os.Exit(1)
	}

	cmd.Help()
	os.Exit(1)
}

func HelpHelp() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "    delphi help [command]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands are:")
	for _, cmd := range Commands {
		fmt.Printf("    %8s %s\n", cmd.Name, cmd.ShortDesc)
	}
}

func Help() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "    delphi command [arguments]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands are:")
	for _, cmd := range Commands {
		fmt.Printf("    %-8s %s\n", cmd.Name, cmd.ShortDesc)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Use \"delphi help [command]\" for more information about a command.")
}

func main() {
	Commands = []Command{
		{"test", test.ShortDesc, test.Main, test.Help},
		{"help", "print help about a command", CommandHelp, HelpHelp},
	}

	if len(os.Args) <= 1 {
		Help()
		os.Exit(1)
	}

	cmdname := os.Args[1]
	args := append([]string{os.Args[0]}, os.Args[2:]...)

	cmd := FindCommand(cmdname)
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "No command named %q\n", cmdname)
		fmt.Fprintln(os.Stderr, "")
		Help()
		os.Exit(1)
	}

	cmd.Main(args)
}
