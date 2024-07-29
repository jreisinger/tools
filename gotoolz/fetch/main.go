package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var usage = `Download HTTP resources in parallel.

fetch [flags] url [url ...]`

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}

	f := flag.String("f", "", "file containing URLs, one per line")
	t := flag.Duration("t", 0, "request timeout")
	v := flag.Bool("v", false, "be verbose")
	flag.Parse()

	urls := flag.Args()
	if *f != "" {
		b, err := os.ReadFile(*f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
			os.Exit(1)
		}
		for _, line := range strings.Split(string(b), "\n") {
			if line != "" {
				urls = append(urls, line)
			}
		}
	}
	urls = dedup(urls)

	c := make(chan *httpResource)
	for _, u := range urls {
		go fetch(u, *t, c)
	}
	for range urls {
		resource := <-c
		if resource.err != nil {
			fmt.Fprintf(os.Stderr, "fetch: %v\n", resource.err)
			continue
		}
		if *v {
			fmt.Printf("%s -> %s (%d bytes)\n", resource.url, resource.file, resource.size)
		}
	}
}

func dedup(ss []string) []string {
	var ss2 []string
	seen := make(map[string]bool)
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			ss2 = append(ss2, s)
		}
	}
	return ss2
}

type httpResource struct {
	url  string
	size int64
	file string
	err  error
}

func fetch(url string, timeout time.Duration, c chan<- *httpResource) {
	resource := httpResource{url: url}

	client := http.Client{Timeout: timeout}
	resp, err := client.Get(resource.url)
	if err != nil {
		resource.err = err
		c <- &resource
		return
	}
	defer resp.Body.Close()

	resource.file = path.Base(url)
	f, err := os.Create(resource.file)
	if err != nil {
		resource.err = err
		c <- &resource
		return
	}
	resource.size, err = io.Copy(f, resp.Body)
	if err != nil {
		resource.err = err
		c <- &resource
		return
	}

	c <- &resource
}
