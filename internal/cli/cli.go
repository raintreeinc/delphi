package cli

import "fmt"

func Clear() {
	fmt.Println("\x1b[3;J\x1b[H\x1b[2J")
}

func Printf(fmt string, args ...interface{}) {}
func Infof(fmt string, args ...interface{})  {}
func Errorf(fmt string, args ...interface{}) {}
