package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/mod/modfile"
)

func usage() {
	fmt.Fprintf(os.Stderr, `gorepojack searches a directory recursively for go.mod files. From them, it
extracts dependencies (Go modules) and evaluates whether they are susceptible to
repository hijacking. The evaluation is done by checking the HTTP response
codes.

usage: gorepojack [options]
`)
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	c = flag.Int("c", 10, "concurrent HTTP requests")
	d = flag.String("d", ".", "`directory` to search")
	v = flag.Bool("v", false, "be verbose")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("gorepojack: ")

	flag.Usage = usage
	flag.Parse()

	modfiles, err := findModFiles(*d)
	if err != nil {
		log.Fatal(err)
	}

	var modules []module
	for _, mf := range modfiles {
		deps, err := extractDeps(mf)
		if err != nil {
			log.Fatal(err)
		}
		for _, dep := range deps {
			modules = append(modules, module{
				goModFilePath: mf,
				path:          dep,
				repoURL:       inferRepoURL(dep),
				userURL:       inferUserURL(dep),
			})
		}
	}

	seen := make(map[string]bool)
	limiter := make(chan struct{}, *c)
	var wg sync.WaitGroup
	for _, mod := range modules {
		if seen[mod.repoURL] {
			continue
		}
		seen[mod.repoURL] = true
		limiter <- struct{}{}
		wg.Add(1)
		go func(mod module) {
			evalModRepo(mod, *v)
			wg.Done()
			<-limiter
		}(mod)
	}
	wg.Wait()
}

type module struct {
	goModFilePath string // /home/bill/github.com/ardanlabs/service/go.mod
	path          string // github.com/user/module/pkg
	repoURL       string // https://github.com/user/module
	userURL       string // https://github/com/user
}

func extractDeps(gomod string) ([]string, error) {
	var paths []string
	b, err := os.ReadFile(gomod)
	if err != nil {
		return nil, err
	}
	mf, err := modfile.Parse(gomod, b, nil)
	if err != nil {
		return nil, err
	}
	for _, r := range mf.Require {
		paths = append(paths, r.Mod.Path)
	}
	return paths, nil
}

func findModFiles(dir string) ([]string, error) {
	var gomods []string
	visit := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && filepath.Base(path) == "go.mod" {
			gomods = append(gomods, path)
		}
		return nil
	}
	err := filepath.WalkDir(dir, visit)
	if err != nil {
		return nil, err
	}
	return gomods, nil
}

func evalModRepo(mod module, verbose bool) {
	if mod.repoURL == "" || mod.userURL == "" {
		return
	}
	repoResp, err := getURL(mod.repoURL)
	if err != nil {
		log.Printf("getting %s: %v\n", mod.repoURL, err)
		return
	}
	switch repoResp.StatusCode {
	case http.StatusOK:
		if verbose {
			fmt.Printf("%-5s %d for %s in %s\n", "OK", repoResp.StatusCode, mod.repoURL, mod.goModFilePath)
		}
	case http.StatusMovedPermanently, http.StatusFound:
		fmt.Printf("%-5s %d for %s -> %s in %s\n", "WARN", repoResp.StatusCode, mod.repoURL, repoResp.Header.Get("location"), mod.goModFilePath)
	case http.StatusNotFound:
		userResp, err := getURL(mod.userURL)
		if err != nil {
			log.Printf("getting %s: %v\n", mod.userURL, err)
			return
		}
		switch userResp.StatusCode {
		case http.StatusOK:
			if verbose {
				fmt.Printf("%-5s %d for %s and %d for %s in %s\n", "OK", repoResp.StatusCode, mod.repoURL, userResp.StatusCode, mod.userURL, mod.goModFilePath)
			}
		case http.StatusMovedPermanently, http.StatusFound:
			fmt.Printf("%-5s %d for %s and %d for %s -> %s in %s\n", "WARN", repoResp.StatusCode, mod.repoURL, userResp.StatusCode, mod.userURL, userResp.Header.Get("location"), mod.goModFilePath)
		case http.StatusNotFound:
			fmt.Printf("%-5s %d for %s and %d for %s in %s\n", "WARN", repoResp.StatusCode, mod.repoURL, userResp.StatusCode, mod.userURL, mod.goModFilePath)
		default:
			fmt.Printf("%-5s %d for %s and %d for %s in %s\n", "WARN", repoResp.StatusCode, mod.repoURL, userResp.StatusCode, mod.userURL, mod.goModFilePath)
		}
	default:
		fmt.Printf("%-5s %d for %s in %s\n", "WARN", repoResp.StatusCode, mod.repoURL, mod.goModFilePath)
	}
}

func inferUserURL(repopath string) string {
	if repopath == "" {
		return ""
	}
	parts := strings.Split(repopath, "/")
	if len(parts) < 2 {
		return ""
	}
	host := parts[0]
	user := parts[1]
	return fmt.Sprintf("https://%s/%s", host, user)
}

func inferRepoURL(repopath string) string {
	if repopath == "" {
		return ""
	}
	parts := strings.Split(repopath, "/")
	if len(parts) < 3 {
		return ""
	}
	host := parts[0]
	user := parts[1]
	repo := parts[2]
	return fmt.Sprintf("https://%s/%s/%s", host, user, repo)
}

func getURL(url string) (*http.Response, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, nil
}
