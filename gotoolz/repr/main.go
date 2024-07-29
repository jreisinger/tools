// Repr returns various representations of integer numbers.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
)

type repr struct {
	input string
	dec   string
	bin   string
	oct   string
	hex   string
	char  string
}

func printReprs(reprs []repr) {
	if len(reprs) == 0 {
		return
	}
	const format = "%v\t%v\t%v\t%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "input", "bin", "oct", "dec", "hex", "char")
	fmt.Fprintf(tw, format, "-----", "---", "---", "---", "---", "----")
	for _, r := range reprs {
		fmt.Fprintf(tw, format, r.input, r.bin, r.oct, r.dec, r.hex, r.char)
	}
	tw.Flush()
}

func main() {
	var b = flag.Int("b", 10, "base of input numbers")
	flag.Parse()

	var reprs []repr
	for _, arg := range flag.Args() {
		i, err := strconv.ParseInt(arg, *b, 0)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"repr: parsing %s as int of base %d: %v\n", arg, *b, err,
			)
			os.Exit(1)
		}
		r := repr{
			input: arg,

			// You could have also used strconv.FormatInt.
			bin: fmt.Sprintf("%b", i),
			oct: fmt.Sprintf("%o", i),
			dec: fmt.Sprintf("%d", i),
			hex: fmt.Sprintf("%x", i),

			char: fmt.Sprintf("%q", i),
		}
		reprs = append(reprs, r)
	}
	printReprs(reprs)
}
