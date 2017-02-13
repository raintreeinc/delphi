package main

import (
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"

	"github.com/loov/watchrun/pgroup"
	"github.com/loov/watchrun/watch"
)

var (
	ignore = watch.Globs{false, watch.DefaultIgnore, nil}
	care   = watch.Globs{false, nil, nil}

	interval = flag.Duration("interval", 300*time.Millisecond, "interval to wait between monitoring")
	monitor  = flag.String("monitor", ".", "files/folders/globs to monitor")
	verbose  = flag.Bool("verbose", false, "verbose output")
)

func init() {
	flag.Var(&ignore, "ignore", "ignore files/folders that match these globs")
	flag.Var(&care, "care", "check only changes to files that match these globs")
}

var (
	fmtBuild   = color.New(color.FgBlack, color.BgWhite)
	fmtDefault = color.New()
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.PrintDefaults()
		return
	}

	monitoring := strings.Split(*monitor, ";")
	ignoring := ignore.All()
	caring := care.All()

	watcher := watch.New(
		*interval,
		monitoring,
		ignoring,
		caring,
		true,
	)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		watcher.Stop()
	}()

	build := &Build{}
	build.dpr = args[0]
	build.bin = "bin"
	for range watcher.Changes {
		build.Rerun()
	}
}

type Build struct {
	dpr    string
	bin    string
	tmpbin string

	compile *exec.Cmd
	execute *exec.Cmd
}

func (build *Build) Rerun() {
	ClearScreen()
	Kill(build.compile)

	tmpbin := filepath.Join(build.bin, "tmp")
	os.Mkdir(tmpbin, 0755)

	fmtBuild.Println("Compiling")
	build.compile = Command("dcc32", "-W", build.dpr, "-E"+tmpbin)
	if err := build.compile.Run(); err != nil {
		fmtBuild.Printf("Failed to compile: %v\n", err)
		return
	}
	fmtBuild.Println("Compiled")

	exe := filepath.Join(build.bin, ChangeExt(filepath.Base(build.dpr), ".exe"))
	tmpexe := filepath.Join(tmpbin, filepath.Base(exe))

	fmtBuild.Printf("Stopping \"%v\"\n", exe)
	if build.execute != nil {
		Kill(build.execute)
	}

	if err := os.Remove(exe); err != nil {
		fmtBuild.Printf("Failed to remove previous executable: %v\n", err)
		return
	}

	if err := os.Rename(tmpexe, exe); err != nil {
		fmtBuild.Printf("Failed to rename build result: %v\n", err)
		return
	}

	fmtBuild.Printf("Starting \"%v\"\n", exe)
	build.execute = Command(exe)
	if err := build.execute.Start(); err != nil {
		fmtBuild.Printf("Failed to start: %v\n", err)
		return
	}
	go func() {
		exe := build.execute
		if err := exe.Wait(); err != nil {
			fmtBuild.Printf("Finished: %v\n", err)
		} else {
			fmtBuild.Printf("Finished\n")
		}
	}()
}

func Command(exe string, args ...string) *exec.Cmd {
	cmd := exec.Command(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	pgroup.Setup(cmd)
	return cmd
}

func Kill(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}

	pgroup.Kill(cmd)
}

func ChangeExt(name string, newext string) string {
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)] + newext
}
