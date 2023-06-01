package markdown

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
)

func ToHTML(markdown []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert(markdown, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Files(fsys fs.FS) ([]string, error) {
	var mdfiles []string
	visit := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if !entry.IsDir() && filepath.Ext(path) == ".md" {
			mdfiles = append(mdfiles, path)
		}
		return nil
	}
	err := fs.WalkDir(fsys, ".", visit)
	return mdfiles, err
}

func ChangeExt(path, ext string) string {
	oldExt := filepath.Ext(path)
	if oldExt != ".md" {
		return path
	}
	bare := strings.TrimSuffix(path, oldExt)
	return bare + ext
}
