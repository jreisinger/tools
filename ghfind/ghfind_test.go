package ghfind

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnmarshalFiles(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "tree.json"))
	if err != nil {
		t.Fatal(err)
	}
	files, err := unmarshalFiles(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("want 2, got %d", len(files))
	}
	if files[0].Path != ".bin" {
		t.Errorf("want .bin, got %s", files[0].Path)
	}
	if files[1].Path != ".bin/check-file-for-starting-slash" {
		t.Errorf("want .bin/check-file-for-starting-slash, got %s", files[0].Path)
	}
}
