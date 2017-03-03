package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	fmtPriority = color.New(color.FgBlack, color.BgWhite)
	fmtError    = color.New(color.FgRed, color.BgWhite)
	fmtDefault  = color.New()
)

func Clear() {
	fmt.Println("\x1b[3;J\x1b[H\x1b[2J")
}

func Printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}

func Warnf(format string, args ...interface{}) {
	fmtError.Fprintf(os.Stdout, format, args...)
}

func Infof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}

func Priorityf(format string, args ...interface{}) {
	fmtPriority.Fprintf(os.Stdout, format, args...)
}

func Helpf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func Errorf(format string, args ...interface{}) {
	fmtError.Fprintf(os.Stderr, format, args...)
}
