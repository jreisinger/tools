package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	ghtml "github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/toc"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("mdsrv: ")

	highlight := flag.Bool("highlight", false, "enable syntax highlighting")
	flag.Parse()

	mdFiles := flag.Args()
	if len(mdFiles) == 0 {
		log.Fatal("please supply *.md file(s)")
	}

	tmpdir, err := os.MkdirTemp("", "mdsrv-*")
	if err != nil {
		log.Fatal(err)
	}

	if err := createStyle(tmpdir); err != nil {
		log.Fatal(err)
	}

	for _, mdfile := range mdFiles {
		// Continually be converting markdown file to html.
		go func(mdfile string) {
			for {
				if err := createHTML(tmpdir, mdfile, *highlight); err != nil {
					log.Fatal(err)
				}
				time.Sleep(time.Second)
			}
		}(mdfile)
	}

	// Start a web server serving content from tmpdir.
	addr := fmt.Sprintf("localhost:%d", 8000)
	log.Printf("serving supplied *.md file(s) from %s at http://%s", tmpdir, addr)
	handler := http.FileServer(http.Dir(tmpdir))
	log.Fatal(http.ListenAndServe(addr, handler))
}

var htmlHead = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <link rel="stylesheet" href="/style.css">
  </head>
  <body>
  <div class="content">
  <!-- Page content -->
`

var htmlTail = `
  </div>
  </body>
</html>
`

// toHTML converts markdown to HTML, adding ToC and optional syntax highlighting.
func toHTML(markdown []byte, highlight bool) ([]byte, error) {
	var buf bytes.Buffer
	var extensions []goldmark.Extender
	extensions = append(extensions, &toc.Extender{Compact: true}, extension.Linkify)
	if highlight {
		extensions = append(extensions, highlighting.Highlighting)
	}
	md := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithRendererOptions(
			// to show images inserted via GitHub web
			ghtml.WithUnsafe(),
		),
	)
	if err := md.Convert(markdown, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// createHTML converts contents of filename from markdown to html, replaces .md
// suffix with .html and stores it in dir keeping the original directory path.
func createHTML(dir string, filename string, highlight bool) error {
	m, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read markdown file: %v", err)
	}

	var h bytes.Buffer
	h.Write([]byte(htmlHead))
	b, err := toHTML(m, highlight)
	if err != nil {
		return err
	}
	h.Write(b)
	h.Write([]byte(htmlTail))

	subdir := filepath.Dir(filename)
	if err := os.MkdirAll(filepath.Join(dir, subdir), 0750); err != nil {
		return err
	}

	name := filepath.Join(dir, strings.TrimSuffix(filename, ".md")+".html")
	if err := os.WriteFile(name, h.Bytes(), 0640); err != nil {
		return fmt.Errorf("write html file: %v", err)
	}
	return nil
}

// createStyle creates style.css in dir.
func createStyle(dir string) error {
	css := `
.content {
	max-width: 960px;
	margin: auto;
}

body {
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji';
	line-height: 1.4;
	font-size: small;
}

h1, h2, h3, h4, h5 { 
	margin-top: 1rem;
	margin-bottom: 1rem;
}

img {
	max-width: 100%;
}

code {
	font-family: monospace;
}

pre {
	background: #f7f7f7;
	border: 1px solid #d7d7d7;
	margin: 1em 1.75em;
	padding: .25em;
	overflow: auto;
	white-space: pre-wrap;
}

blockquote {
	font-family: cursive;
}

@media screen and (max-device-width: 480px) {
	body {
		-webkit-text-size-adjust: none;
	}
}
`
	return os.WriteFile(filepath.Join(dir, "style.css"), []byte(css), 0640)
}
