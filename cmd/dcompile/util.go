package main

import "fmt"

func ClearScreen() {
	fmt.Println("\x1b[3;J\x1b[H\x1b[2J")
}
