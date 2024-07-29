package main

import (
	"flag"
	"fmt"
	"os"
	"recent"
)

var defaultDir string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "recent: get working directory: %v\n", err)
		os.Exit(1)
	}
	defaultDir = wd
}

func main() {
	d := flag.String("d", defaultDir, "directory to search")
	e := flag.String("e", "", "exclude paths matching regexp")
	n := flag.Int("n", 10, "number of files")
	flag.Parse()

	files, err := recent.Files(os.DirFS(*d), *n, *e)
	if err != nil {
		fmt.Fprintf(os.Stderr, "recent: searching %s: %v\n", *d, err)
		os.Exit(1)
	}

	for _, f := range files {
		fmt.Printf("%s\t%s\n",
			f.ModTime.Local().Format("2006-01-02 15:04:05"),
			f.Path,
		)
	}
}
