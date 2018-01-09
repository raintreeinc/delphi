package test

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/raintreeinc/delphi/delphi"
	"github.com/raintreeinc/delphi/internal/cli"
	"github.com/raintreeinc/delphi/internal/walk"
	"github.com/raintreeinc/delphi/scanner"
	"github.com/raintreeinc/delphi/token"
)

const ShortDesc = "test units"

func Help(args []string) {
	cli.Helpf("Usage:\n")
	cli.Helpf("\t%s [filename]\n\n", args[0])
	cli.Helpf(`Arguments:
  -build    build directory
  -search   search path
  -define   compilator defines
  -root     search path root, add all folders recursively

  -dunit    generate DUnit tests
`)
}

type Flags struct {
	Help    bool
	Verbose bool

	BuildDir string
	Search   string
	Root     string
	Define   string
	Paths    []string

	DUnit string

	Set *flag.FlagSet
}

func (flags *Flags) Parse(args []string) {
	flags.Set = flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.Set.BoolVar(&flags.Help, "help", false, "show help")
	flags.Set.BoolVar(&flags.Help, "h", false, "show help")

	flags.Set.BoolVar(&flags.Verbose, "v", false, "verbose")
	flags.Set.BoolVar(&flags.Verbose, "verbose", false, "verbose")

	flags.Set.StringVar(&flags.BuildDir, "build", "", "build directory, default DELPHI_TEMP")
	flags.Set.StringVar(&flags.Search, "search", "", "search path, default DELPHI_SEARCH")
	flags.Set.StringVar(&flags.Define, "define", "", "compile defines, default DELPHI_DEFINE")
	flags.Set.StringVar(&flags.Root, "root", "", "search root, adds all folders recursively")

	flags.Set.StringVar(&flags.DUnit, "dunit", "", "generate DUnit tests")

	flags.Set.Parse(args[1:])

	flags.Paths = flags.Set.Args()

	// convert to absolute paths
	flags.BuildDir = delphi.AbsPath(flags.BuildDir)
	flags.Search = delphi.AbsPath(flags.Search)
	flags.Root = delphi.AbsPath(flags.Root)
	flags.Paths = delphi.AbsPaths(flags.Paths)
}

func cleanup(build *Build, tempBuildDir string) {
	build.Kill()
	if tempBuildDir != "" {
		os.RemoveAll(tempBuildDir)
	}
}

func signalhandler(build *Build, tempBuildDir string) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	for range signals {
		cleanup(build, tempBuildDir)
		os.Exit(1)
		break
	}
}

func Main(args []string) {
	var flags Flags
	flags.Parse(args)
	if flags.Help || len(flags.Paths) == 0 {
		Help(args)
		return
	}

	var tempdir string
	if flags.BuildDir == "" {
		tempdir, _ = ioutil.TempDir(delphi.TempDir(), "delphitest")
		flags.BuildDir = tempdir
	}
	if flags.Search == "" {
		flags.Search = delphi.SearchPath()
	}
	build := &Build{}
	defer cleanup(build, tempdir)

	build.Verbose = flags.Verbose
	build.Name = "All"
	build.Dir = flags.BuildDir
	build.Project = build.Name + "_Tests"
	build.Search = strings.Split(flags.Search, ";")
	build.Define = strings.Split(flags.Define, ";")

	if flags.Root != "" {
		for _, p := range delphi.SearchPathFromRoot(flags.Root) {
			if !contains(p, build.Search) {
				build.Search = append(build.Search, p)
			}
		}
	}

	if err := build.Prepare(); err != nil {
		cli.Errorf("%v\n", err)
		return
	}
	go signalhandler(build, tempdir)

	cli.Priorityf("collecting tests\n")

	filenames := make(chan string, 8)
	errors := make(chan error, 8)
	go func() {
		walk.Globs(flags.Paths, filenames, errors, walk.IsDelphiFile)
		close(filenames)
		close(errors)
	}()
	go func() {
		for err := range errors {
			if err != nil {
				cli.Errorf("%v\n", err)
			}
		}
	}()

	for filename := range filenames {
		if !strings.EqualFold(filepath.Ext(filename), ".pas") {
			continue
		}

		test, err := NewTestFile(filename)
		if err != nil {
			cli.Errorf("%v\n", err)
			continue
		}
		if len(test.Funcs) == 0 {
			continue
		}

		dir := filepath.Dir(filename)
		if !contains(dir, build.Search) {
			build.Search = append(build.Search, dir)
		}
		build.Tests = append(build.Tests, test)
	}

	if flags.Verbose {
		cli.Infof("Building: %v\n", build.Project)
		cli.Infof("     DPR: %v\n", build.DPR())
		cli.Infof("     Dir: %v\n", build.Dir)

		cli.Infof("  Search:\n")
		for _, path := range build.Search {
			cli.Infof("    %v\n", path)
		}

		cli.Infof("  Define:\n")
		for _, define := range build.Define {
			cli.Infof("    %v\n", define)
		}
	}

	if flags.DUnit != "" {
		cli.Infof("Generating DUnit for:\n")
		for _, testfile := range build.Tests {
			cli.Infof("    %v\n", testfile.UnitName)
			for _, testname := range testfile.Funcs {
				cli.Infof("        %v\n", testname)
			}
		}
		if err := GenerateDUnit(build.Tests, flags.DUnit); err != nil {
			cli.Errorf("%v\n", err)
		}
		return
	}

	if err := build.Create(); err != nil {
		cli.Errorf("%v\n", err)
		return
	}

	if err := build.Run(); err != nil {
		cli.Errorf("%v\n", err)
		return
	}
}

type Build struct {
	Name    string
	Project string
	Verbose bool

	Dir    string
	Define []string
	Search []string

	Tests []*TestFile

	Compile *exec.Cmd
	Execute *exec.Cmd
}

func (build *Build) DPR() string { return filepath.Join(build.Dir, build.Project+".dpr") }
func (build *Build) CFG() string { return filepath.Join(build.Dir, build.Project+".cfg") }
func (build *Build) DOF() string { return filepath.Join(build.Dir, build.Project+".dof") }

func (build *Build) EXE() string { return filepath.Join(build.OutputDir(), build.Project+".exe") }

func (build *Build) OutputDir() string { return filepath.Join(build.Dir, build.Project+"_bin") }
func (build *Build) BuildDir() string  { return filepath.Join(build.Dir, build.Project+"_dcu") }

func (build *Build) Kill() error {
	var err1, err2 error
	if build.Compile.Process != nil {
		err1 = build.Compile.Process.Kill()
	}
	if build.Execute.Process != nil {
		err2 = build.Execute.Process.Kill()
	}
	return NewErrors("killing build", err1, err2)
}

func (build *Build) Prepare() error {
	// prepare folders
	err := NewErrors("prepare",
		os.MkdirAll(build.OutputDir(), 0755),
		os.MkdirAll(build.BuildDir(), 0755),
	)
	if err != nil {
		return err
	}

	build.Compile = exec.Command(delphi.DCC(), build.DPR())
	if build.Verbose {
		build.Execute = exec.Command(build.EXE(), "-v")
	} else {
		build.Execute = exec.Command(build.EXE())
	}

	build.Execute.Stdout = os.Stdout
	build.Compile.Stdout = os.Stdout

	return nil
}

func (build *Build) Create() error {
	return NewErrors("create",
		CreateFile(build.DPR(), DPR_Template, build),
		CreateFile(build.DOF(), DOF_Template, build),
		CreateFile(build.CFG(), CFG_Template, build),
	)
}

func (build *Build) Run() error {
	cli.Priorityf("running compiler\n")
	if err := build.Compile.Run(); err != nil {
		return err
	}
	cli.Priorityf("running tests\n")
	if err := build.Execute.Run(); err != nil {
		return err
	}
	return nil
}

type TestFile struct {
	Path     string
	Full     string
	UnitName string
	Funcs    []string
}

func NewTestFile(path string) (*TestFile, error) {
	ext := filepath.Ext(path)
	file := &TestFile{
		Path:     path,
		UnitName: filepath.Base(path[:len(path)-len(ext)]),
		Funcs:    []string{},
	}

	file.Path = delphi.AbsPath(path)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pre token.Token
	scanner.Scan(data, 0, func(tok token.Token, lit string) error {
		if pre == token.PROCEDURE && tok == token.IDENT {
			if strings.HasPrefix(strings.ToLower(lit), "test_") {
				if !contains(lit, file.Funcs) {
					file.Funcs = append(file.Funcs, lit)
				}
			}
		}
		pre = tok
		return nil
	}, nil)

	return file, nil
}
