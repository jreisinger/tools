/*
Mdsrv converts supplied Markdown files or all Markdown files in current
directory to temporary HTML files and serves them over HTTP.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/jreisinger/tools/internal/mdsrv/html"
	"github.com/jreisinger/tools/internal/mdsrv/markdown"

	cp "github.com/otiai10/copy"
)

var (
	p = flag.Int("p", 8000, "port")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("mdsrv: ")
	flag.Parse()

	if len(flag.Args()) != 1 || !markdown.Is(flag.Args()[0]) {
		log.Fatal("please supply .md file")
	}
	mdfile := flag.Args()[0]

	tmpdir, err := tmpSubdir("/tmp")
	if err != nil {
		log.Fatal(err)
	}

	if err := html.AddCSS(tmpdir); err != nil {
		log.Fatal(err)
	}

	if err := copyStatic(tmpdir); err != nil {
		log.Fatal(err)
	}

	// Continually be converting markdown file to html.
	go func() {
		for {
			if err := markdown.ToHTML(tmpdir, mdfile); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second)
		}
	}()

	addr := fmt.Sprintf("localhost:%d", *p)
	htmlfile := markdown.ChangeExt(mdfile, ".html")
	log.Printf("serving file from %s at http://%s/%s", tmpdir, addr, htmlfile)
	handler := http.FileServer(http.Dir(tmpdir))
	log.Fatal(http.ListenAndServe(addr, handler))
}

// copyStatic copies static folder, if it exists, to tmpdir recursively.
func copyStatic(tmpdir string) error {
	if _, err := os.Stat("static"); os.IsNotExist(err) {
		return nil
	}
	return cp.Copy("static", filepath.Join(tmpdir, "static"))
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
