package uses

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/raintreeinc/delphi/delphi"
	"github.com/raintreeinc/delphi/internal/cli"
)

const ShortDesc = "print unit uses graph"

func Help(args []string) {
	cli.Helpf("Usage:\n")
	cli.Helpf("\t%s project.dpr\n\n", args[0])
	cli.Helpf(`Arguments:
  -search    search path
  -root      search path root, add all folders recursively

  -out       output file

  -why       why is a particular file included
  -interface only analyse interface section
`)
}

type Flags struct {
	Help    bool
	Verbose bool

	Search string
	Root   string
	Output string

	Paths []string

	Why string

	InterfaceOnly bool

	Set *flag.FlagSet
}

func (flags *Flags) Parse(args []string) {
	flags.Set = flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.Set.BoolVar(&flags.Help, "help", false, "show help")
	flags.Set.BoolVar(&flags.Help, "h", false, "show help")

	flags.Set.BoolVar(&flags.Verbose, "v", false, "verbose")
	flags.Set.BoolVar(&flags.Verbose, "verbose", false, "verbose")

	flags.Set.StringVar(&flags.Search, "search", "", "search path, default DELPHI_SEARCH")
	flags.Set.StringVar(&flags.Root, "root", "", "search path root, add all folders recursively")

	flags.Set.StringVar(&flags.Output, "out", "", "output file")
	flags.Set.StringVar(&flags.Why, "why", "", "why is a particular file included")

	flags.Set.BoolVar(&flags.InterfaceOnly, "interface", false, "only scan interfaces")

	flags.Set.Parse(args[1:])
	flags.Paths = flags.Set.Args()
}

func Main(args []string) {
	var flags Flags
	flags.Parse(args)
	if flags.Help || len(flags.Paths) == 0 {
		Help(args)
		return
	}

	if flags.Search == "" {
		flags.Search = delphi.SearchPath()
	}

	index := NewIndex()

	if flags.Root != "" {
		index.AddSourceDir(flags.Root)
	}
	if flags.Search != "" {
		for _, p := range strings.Split(flags.Search, ";") {
			index.AddSourceDir(p)
		}
	}

	index.Verbose = flags.Verbose
	index.Build(flags.Paths)

	if flags.Why != "" {
		for _, reason := range Why(index, flags.Why) {
			fmt.Println(reason)
		}
		return
	}

	if flags.Output == "" {
		flags.Output = trimExt(filepath.Base(flags.Paths[0])) + ".txt"
	}

	file, err := os.Create(flags.Output)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	wr := bufio.NewWriter(file)
	defer wr.Flush()

	ext := strings.ToLower(filepath.Ext(flags.Output))
	if ext == ".tgf" {
		WriteTGF(index, wr)
	} else if ext == ".dot" {
		WriteDOT(index, wr)
	} else if ext == ".txt" {
		WriteTXT(index, wr)
	} else if ext == ".glay" {
		WriteGLAY(index, wr)
	} else {
		log.Fatal("Unknown file extension " + ext)
	}
}
