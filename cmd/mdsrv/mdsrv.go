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
	"bytes"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/jreisinger/tools/html"
	"github.com/jreisinger/tools/markdown"
)

func main() {
	var mdfiles []string

	if len(os.Args[1:]) > 0 {
		mdfiles = os.Args[1:]
	} else {
		var err error
		if mdfiles, err = markdown.Files(os.DirFS(".")); err != nil {
			log.Fatal(err)
		}
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

		var buf bytes.Buffer
		buf.Write([]byte(html.Head))
		h, err := markdown.ToHTML(m)
		if err != nil {
			log.Fatal(err)
		}
		buf.Write(h)
		buf.Write([]byte(html.Tail))

		dir := filepath.Dir(mdfile)
		if err := os.MkdirAll(filepath.Join(tmpdir, dir), 0750); err != nil {
			log.Fatal(err)
		}

		htmlfile := markdown.ChangeExt(mdfile, ".html")
		if err := os.WriteFile(
			filepath.Join(tmpdir, htmlfile), buf.Bytes(), 0640); err != nil {
			log.Fatalf("write html file: %v", err)
		}
	}

	if err := html.AddCSS(tmpdir); err != nil {
		log.Fatal(err)
	}

	addr := "localhost:8000"
	log.Printf("serving %d file(s) from %s at %s", len(mdfiles), tmpdir, addr)
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
