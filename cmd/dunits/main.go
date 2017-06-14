package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	outfile   = flag.String("out", "", "output file")
	directory = flag.String("dir", "", "folder to search for units")
	verbose   = flag.Bool("v", false, "verbose output")

	interfaceOnly = flag.Bool("interface", false, "only output uses in interface")
)

func main() {
	flag.Parse()

	rootfiles := flag.Args()
	if len(rootfiles) == 0 {
		log.Println("dpr not specified")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *directory == "" {
		*directory = filepath.Dir(rootfiles[0])
	}

	index, err := NewIndex(*directory)
	if err != nil {
		log.Fatal(err)
	}

	index.Build(rootfiles)

	if *outfile == "" {
		*outfile = TrimExt(filepath.Base(rootfiles[0])) + ".txt"
	}

	file, err := os.Create(*outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	wr := bufio.NewWriter(file)
	defer wr.Flush()

	ext := strings.ToLower(filepath.Ext(*outfile))
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
