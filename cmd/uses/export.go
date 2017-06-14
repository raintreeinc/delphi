package uses

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func WriteTXT(index *Index, out io.Writer) (n int, err error) {
	write := func(format string, args ...interface{}) bool {
		if err != nil {
			return false
		}
		var x int
		x, err = fmt.Fprintf(out, format, args...)
		n += x
		return err == nil
	}

	if cycles := FindCycles(index); len(cycles) > 0 {
		write("Circular interface uses:\n")
		for _, cycle := range cycles {
			write("\t%v\n", cycle)
		}
		write("\n")
	}

	cunitnames := make([]string, 0, len(index.Uses))
	for cunitname := range index.Uses {
		cunitnames = append(cunitnames, cunitname)
	}
	sort.Strings(cunitnames)

	for _, cunitname := range cunitnames {
		uses := index.Uses[cunitname]

		write("# %v\n", uses.Unit)
		for _, use := range uses.Interface {
			write("\t+ %v\n", index.NormalName(use))
		}
		for _, use := range uses.Implementation {
			write("\t- %v\n", index.NormalName(use))
		}
		write("\n")
	}

	return
}

func WriteDOT(index *Index, out io.Writer) (n int, err error) {
	write := func(format string, args ...interface{}) bool {
		if err != nil {
			return false
		}
		var x int
		x, err = fmt.Fprintf(out, format, args...)
		n += x
		return err == nil
	}

	write("digraph G{\n")
	for _, uses := range index.Uses {
		for _, use := range uses.Interface {
			write("\t%v -> %v;\n", uses.Unit, index.NormalName(use))
		}
		for _, use := range uses.Implementation {
			write("\t%v -> %v [style=dashed;dir=both;weight=0];\n", uses.Unit, index.NormalName(use))
		}
	}
	write("}\n")

	return
}

func WriteTGF(index *Index, out io.Writer) (n int, err error) {
	write := func(format string, args ...interface{}) bool {
		if err != nil {
			return false
		}
		var x int
		x, err = fmt.Fprintf(out, format, args...)
		n += x
		return err == nil
	}

	ids := make(map[string]int, len(index.Uses))

	id := 0
	for cunitname, use := range index.Uses {
		id++
		ids[cunitname] = id

		write("%v %v\n", id, use.Unit)
	}

	write("#\n")

	for cunitname, uses := range index.Uses {
		for _, use := range uses.Interface {
			write("%v %v\n", ids[cunitname], ids[strings.ToLower(use)])
		}
		for _, use := range uses.Implementation {
			write("%v %v\n", ids[cunitname], ids[strings.ToLower(use)])
		}
	}

	return 0, nil
}
func WriteGLAY(index *Index, out io.Writer) (n int, err error) {
	write := func(format string, args ...interface{}) bool {
		if err != nil {
			return false
		}
		var x int
		x, err = fmt.Fprintf(out, format, args...)
		n += x
		return err == nil
	}

	for _, uses := range index.Uses {
		for _, use := range uses.Interface {
			write("\t%v -> %v;\n", uses.Unit, index.NormalName(use))
		}
		for _, use := range uses.Implementation {
			write("\t%v -> %v;\n", uses.Unit, index.NormalName(use))
		}
	}

	return 0, nil
}
