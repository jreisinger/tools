package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("helm-hardcoded: ")

	flagChart := flag.String("chart", ".", "helm chart to search for hardcoded value; folder or .tgz")
	flagVerbose := flag.Bool("verbose", false, "print also lines containing value")
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalf("supply value to search")
	}
	value := flag.Args()[0]

	loadedChart, err := loader.Load(*flagChart)
	if err != nil {
		log.Fatalf("loading chart: %v", err)
	}

	var allCharts []*chart.Chart
	allCharts = append(allCharts, loadedChart)
	allCharts = append(allCharts, loadedChart.Dependencies()...)

	for _, ch := range allCharts {
		for _, tpl := range ch.Templates {
			lines := linesWithHardcodedValue(string(tpl.Data), value)
			n := len(lines)
			if n > 0 {
				tplPath := filepath.Join(ch.ChartFullPath(), tpl.Name)
				fmt.Printf("%s (%d %s)\n", removeFirstStringBeforeSlash(tplPath), n, formatLines(n))
				if *flagVerbose {
					for _, line := range lines {
						fmt.Println(line)
					}
				}
			}
		}
	}
}

func formatLines(count int) string {
	if count == 1 {
		return "line"
	}
	return "lines"
}

func removeFirstStringBeforeSlash(path string) string {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return path
}

func linesWithHardcodedValue(templateContent, value string) []string {
	var results []string
	lines := strings.Split(templateContent, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.Contains(trimmedLine, "{{") && strings.Contains(trimmedLine, "}}") {
			continue
		}
		if strings.Contains(trimmedLine, value) {
			results = append(results, line)
		}
	}
	return results
}
