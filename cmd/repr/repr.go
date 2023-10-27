/*
Repr returns various representations of a number
- binary
- octal
- hex
*/
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
}

func printReprs(reprs []repr) {
	const format = "%v\t%v\t%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "input", "bin", "oct", "dec", "hex")
	fmt.Fprintf(tw, format, "-----", "---", "---", "---", "---")
	for _, r := range reprs {
		fmt.Fprintf(tw, format, r.input, r.bin, r.oct, r.dec, r.hex)
	}
	tw.Flush()
}

func main() {
	var base = flag.Int("b", 10, "base of input numbers")
	flag.Parse()

	var reprs []repr
	for _, arg := range flag.Args() {
		i, err := strconv.ParseInt(arg, *base, 0)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"repr: converting %s to int of base %d: %v\n", arg, *base, err,
			)
			os.Exit(1)
		}
		r := repr{
			input: arg,
			bin:   strconv.FormatInt(i, 2),
			oct:   strconv.FormatInt(i, 8),
			dec:   strconv.FormatInt(i, 10),
			hex:   strconv.FormatInt(i, 16),
		}
		reprs = append(reprs, r)
	}
	printReprs(reprs)
}
