package markdown

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jreisinger/tools/internal/mdsrv/html"
	"github.com/yuin/goldmark"
	ghtml "github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/toc"
)

// toHTML converts markdown to HTML, adding ToC and syntax highlighting.
func toHTML(markdown []byte) ([]byte, error) {
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(
			&toc.Extender{Compact: true},
			// highlighting.Highlighting,
		),
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

// Files walks fsys recursively and returns found markdown files.
func Files(fsys fs.FS) ([]string, error) {
	var mdfiles []string
	visit := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if !entry.IsDir() && Is(path) {
			mdfiles = append(mdfiles, path)
		}
		return nil
	}
	err := fs.WalkDir(fsys, ".", visit)
	return mdfiles, err
}

func Is(file string) bool {
	return filepath.Ext(file) == ".md"
}

func ChangeExt(path, ext string) string {
	oldExt := filepath.Ext(path)
	if oldExt != ".md" {
		return path
	}
	bare := strings.TrimSuffix(path, oldExt)
	return bare + ext
}

// ToHTML converts mdfile to html file, changes the extension from .md to .html
// and stores it in dir keeping the original directory path.
func ToHTML(dir string, mdfile string) error {
	m, err := os.ReadFile(mdfile)
	if err != nil {
		return fmt.Errorf("read markdown file: %v", err)
	}

	var h bytes.Buffer
	h.Write([]byte(html.Head))
	b, err := toHTML(m)
	if err != nil {
		return err
	}
	h.Write(b)
	h.Write([]byte(html.Tail))

	subdir := filepath.Dir(mdfile)
	if err := os.MkdirAll(filepath.Join(dir, subdir), 0750); err != nil {
		return err
	}

	htmlfile := ChangeExt(mdfile, ".html")
	if err := os.WriteFile(
		filepath.Join(dir, htmlfile), h.Bytes(), 0640); err != nil {
		return fmt.Errorf("write html file: %v", err)
	}
	return nil
}
