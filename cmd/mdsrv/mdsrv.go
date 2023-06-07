/*
Mdsrv converts supplied Markdown files or all Markdown files in current
directory to temporary HTML files and serves them over HTTP.

TODO
  - [x] covert MD to temporary HTML using goldmark
  - [x] use http.FileServer to server HTML files
  - [x] if a file is changed or a new file is created re-render them
*/
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jreisinger/tools/html"
	"github.com/jreisinger/tools/markdown"
)

func main() {
	tmpdir, err := tmpSubdir("/tmp")
	if err != nil {
		log.Fatal(err)
	}

	if err := html.AddCSS(tmpdir); err != nil {
		log.Fatal(err)
	}

	// Continually be converting markdown files to html.
	go func() {
		for {
			mdfiles, err := getMDfiles(os.Args[1:])
			if err != nil {
				log.Fatal(err)
			}
			if err := markdown.ToHTML(tmpdir, mdfiles); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second)
		}
	}()

	addr := "localhost:8000"
	log.Printf("serving files from %s at http://%s", tmpdir, addr)
	handler := http.FileServer(http.Dir(tmpdir))
	log.Fatal(http.ListenAndServe(addr, handler))
}

// getMDfiles filters out markdown files from CLI arguments. If there are no CLI
// arguments it searches current directory recursively.
func getMDfiles(CLIargs []string) (mdfiles []string, err error) {
	if len(CLIargs) > 0 {
		for _, arg := range CLIargs {
			if markdown.Is(arg) {
				mdfiles = append(mdfiles, arg)
			}
		}
		return
	}
	return markdown.Files(os.DirFS("."))
}

// tmpSubdir creates a temporary subdir in dir prefixed with mdsrv. It gets
// removed on Ctrl-C.
func tmpSubdir(dir string) (string, error) {
	tmpdir, err := os.MkdirTemp(dir, "mdsrv")
	if err != nil {
		return "", err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go remove(tmpdir, c)

	return tmpdir, nil
}

// remove removes dir recursively when it receives a signal and then exits.
func remove(dir string, c <-chan os.Signal) {
	<-c
	log.Printf("removing %s", dir)
	if err := os.RemoveAll(dir); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
