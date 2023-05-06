// Ghstats provides statistics about GitHub repositories.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/jreisinger/tools/ghstats"
)

var (
	n = flag.Int("n", 10, "show top `N` repositories")
	c = flag.Int("c", 3, "sort by column number `N`")
)

func main() {
	log.SetPrefix(os.Args[0] + ": ")
	log.SetFlags(0)

	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalf("supply GitHub username")
	}
	user := flag.Args()[0]

	token := os.Getenv("GH_TOKEN")
	if token == "" {
		log.Fatalf("set GH_TOKEN environment variable")
	}

	stats, err := ghstats.Get(user, token)
	if err != nil {
		log.Fatal(err)
	}

	stats.Sort(*c)
	stats.Print(*n)
}
