/*
Mdsrv converts Markdown files in current directory to temporary HTML files and
serves them over HTTP.

TODO
  - [x] covert MD to temporary HTML using goldmark
  - [x] use http.FileServer to server HTML files
  - [x] if a file is changed or a new file is created re-render them
  - [ ] if a file is removed, remove it from tmpdir
*/
package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/jreisinger/tools/html"
	"github.com/jreisinger/tools/markdown"
)

func main() {
	tmpdir, err := os.MkdirTemp("/tmp", "mdsrv")
	if err != nil {
		log.Fatal(err)
	}

	// Cleanup on Ctrl-C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go cleanup(tmpdir, c)

	go func() {
		for {
			mdfiles, err := getMDfiles(os.Args[1:])
			if err != nil {
				log.Fatal(err)
			}
			if err := genHTMLfiles(tmpdir, mdfiles); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second)
		}
	}()

	if err := html.AddCSS(tmpdir); err != nil {
		log.Fatal(err)
	}

	addr := "localhost:8000"
	log.Printf("serving files from %s at %s", tmpdir, addr)
	handler := http.FileServer(http.Dir(tmpdir))
	log.Fatal(http.ListenAndServe(addr, handler))
}

func getMDfiles(CLIargs []string) (mdfiles []string, err error) {
	if len(CLIargs) <= 0 {
		mdfiles, err = markdown.Files(os.DirFS("."))
	} else {
		mdfiles = CLIargs
	}
	return
}

func genHTMLfiles(dir string, mdfiles []string) error {
	for _, mdfile := range mdfiles {
		m, err := os.ReadFile(mdfile)
		if err != nil {
			return fmt.Errorf("read markdown file: %v", err)
		}

		var h bytes.Buffer
		h.Write([]byte(html.Head))
		b, err := markdown.ToHTML(m)
		if err != nil {
			return err
		}
		h.Write(b)
		h.Write([]byte(html.Tail))

		subdir := filepath.Dir(mdfile)
		if err := os.MkdirAll(filepath.Join(dir, subdir), 0750); err != nil {
			return err
		}

		htmlfile := markdown.ChangeExt(mdfile, ".html")
		if err := os.WriteFile(
			filepath.Join(dir, htmlfile), h.Bytes(), 0640); err != nil {
			return fmt.Errorf("write html file: %v", err)
		}
	}
	return nil
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
