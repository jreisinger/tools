// Ghfind searches files in a GitHub repository.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jreisinger/tools/ghfind"
)

var (
	branch = flag.String("branch", "main", "git branch")
	regex  = flag.String("regex", "", "regex to match in file paths (case insensitive)")
)

func main() {
	flag.Usage = func() {
		desc := "Search files in a GitHub repository."
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\n%s [flags] <owner>/<repo>\n", desc, os.Args[0])
		flag.PrintDefaults()
	}

	// Parse CLI arguments.
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}
	repo := flag.Args()[0]

	// Set CLI-style logging.
	log.SetFlags(0)
	log.SetPrefix(os.Args[0] + ": ")

	files, err := ghfind.Files(repo, *branch)
	if err != nil {
		log.Fatal(err)
	}
	ghfind.Print(files, *regex)
}
