package test

import "fmt"

const ShortDesc = "test units"

func Help(args []string) {
	fmt.Println("Usage:")
	fmt.Printf("\t%s [arguments]\n", args[0])
}

func Main(args []string) {
	if len(args) <= 1 {
		Help(args)
		return
	}
}
