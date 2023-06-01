/*
Mdsrv converts Markdown files in current directory to temporary HTML files and
serves them over HTTP.

TODO
  - [x] covert MD to temporary HTML using goldmark
  - [x] use http.FileServer to server HTML files
  - [ ] if a file is changed or a new file is created re-render them
  - [ ] pull remote repo
*/
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/jreisinger/tools/markdown"
)

const dir = "."

func main() {
	mdfiles, err := markdown.Files(os.DirFS(dir))
	if err != nil {
		log.Fatal(err)
	}

	tmpdir, err := os.MkdirTemp("/tmp", "mdsrv")
	if err != nil {
		log.Fatal(err)
	}

	// Cleanup on Ctrl-C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go cleanup(tmpdir, c)

	for _, mdfile := range mdfiles {
		m, err := os.ReadFile(mdfile)
		if err != nil {
			log.Fatalf("read markdown file: %v", err)
		}
		h, err := markdown.ToHTML(m)
		if err != nil {
			log.Fatal(err)
		}

		dir := filepath.Dir(mdfile)
		if err := os.MkdirAll(filepath.Join(tmpdir, dir), 0750); err != nil {
			log.Fatal(err)
		}

		htmlfile := markdown.ChangeExt(mdfile, ".html")
		if err := os.WriteFile(
			filepath.Join(tmpdir, htmlfile), h, 0640); err != nil {
			log.Fatalf("write html file: %v", err)
		}
	}

	addr := "localhost:8000"
	log.Printf("serving %d files from %s at %s", len(mdfiles), tmpdir, addr)
	handler := http.FileServer(http.Dir(tmpdir))
	log.Fatal(http.ListenAndServe(addr, handler))
}

func cleanup(dir string, c <-chan os.Signal) {
	<-c
	log.Printf("clean up %s", dir)
	err := os.RemoveAll(dir)
	if err != nil {
		log.Printf("clean up %s: %v", dir, err)
	}
	os.Exit(0)
}