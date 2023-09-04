// Cncf-projects finds out what programming languages are used for CNCF projects.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jreisinger/tools"
	"github.com/jreisinger/tools/cncf"
)

func main() {
	log.SetPrefix("cncflang: ")
	log.SetFlags(0)

	if len(os.Args[1:]) != 1 {
		log.Fatalf("supply CSV file downloaded from https://landscape.cncf.io/card-mode?project=hosted")
	}
	csvFile := os.Args[1]

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatalf("set GITHUB_TOKEN environment variable")
	}

	projects, err := cncf.CSV(csvFile, githubToken)
	if err != nil {
		log.Fatal(err)
	}

	// Languages stores number of projects per language.
	projectsPerLanguage := make(map[string]int)

	for _, project := range projects {
		if project.Err != nil {
			log.Printf("%s: %v", project.GithubURL, project.Err)
		}
		if _, ok := projectsPerLanguage[project.TopLanguage]; !ok {
			projectsPerLanguage[project.TopLanguage] = 1
		} else {
			projectsPerLanguage[project.TopLanguage] += 1
		}
	}

	var totalProjects int
	for _, l := range tools.SortMapByValue(projectsPerLanguage, false) {
		totalProjects += l.Value
	}
	for _, l := range tools.SortMapByValue(projectsPerLanguage, true) {
		perc := float64(l.Value) / float64(totalProjects) * 100
		fmt.Printf("%3d (%2.0f%%) %s\n", l.Value, perc, l.Key)
	}
	fmt.Println("---")
	fmt.Println(totalProjects)
}
