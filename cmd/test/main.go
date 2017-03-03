package test

import "fmt"

const ShortDesc = "test units"

func Help() {
	fmt.Println("Usage:")
	fmt.Println("    delphi test [arguments]")
}

func Main(args []string) {
	if len(args) <= 1 {
		Help()
		return
	}
}
