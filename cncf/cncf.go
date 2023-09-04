package cncf

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/jreisinger/tools"
	"golang.org/x/oauth2"
)

type Project struct {
	GithubURL   string
	TopLanguage string
	Err         error
}

// CSV extracts GitHub URLs of CNCF projects from CSV file downloaded from
// https://landscape.cncf.io/card-mode?project=hosted. Then it retrieves the
// most used programming language for each project using GitHub's API.
func CSV(csvFile, githubToken string) ([]Project, error) {
	urls, err := githubURLs(csvFile)
	if err != nil {
		return nil, err
	}

	c := make(chan Project)
	for _, url := range urls {
		go func(url string) {
			lang, err := topLanguage(url, githubToken)
			c <- Project{GithubURL: url, Err: err, TopLanguage: lang}
		}(url)
	}
	var projects []Project
	for range urls {
		project := <-c
		if project.Err != nil {
			return nil, project.Err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func githubURLs(csvfile string) ([]string, error) {
	file, err := os.Open(csvfile)
	if err != nil {
		return nil, err
	}
	var repos []string
	r := csv.NewReader(file)
	if _, err := r.Read(); err != nil { // skip first line, the column name
		return nil, err
	}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) < 29 {
			return nil, fmt.Errorf("no column 29 in %s", csvfile)
		}
		repos = append(repos, record[29])
	}
	return repos, nil
}

// topLanguage returns the most used programming topLanguage for code in repo.
func topLanguage(repoURL, githubToken string) (string, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	c := oauth2.NewClient(context.TODO(), ts)
	client := github.NewClient(c)
	owner, repo := getOwnerAndRepo(repoURL)
	languages, _, err := client.Repositories.ListLanguages(context.TODO(), owner, repo)
	if err != nil {
		return "", err
	}
	if len(languages) == 0 {
		return "", nil
	}
	l := tools.SortMapByValue(languages, true)
	return l[0].Key, nil
}

func getOwnerAndRepo(repoURL string) (string, string) {
	// https://github.com/containerd/containerd
	fields := strings.Split(repoURL, "/")
	return fields[len(fields)-2], fields[len(fields)-1]
}
