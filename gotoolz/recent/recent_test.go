package recent_test

import (
	"testing"
	"testing/fstest"
	"time"

	"recent"
)

func TestFilesReturnsTwoFiles(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"file1.txt": {},
		"file2.txt": {},
		"file3.txt": {},
	}
	files, err := recent.Files(fsys, 2, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("wanted %d, got %d files", 2, len(files))
	}
}

func TestFilesSortsFilesByModTime(t *testing.T) {
	t.Parallel()
	now := time.Now()
	fsys := fstest.MapFS{
		"youngest.txt": &fstest.MapFile{ModTime: now.Add(-1 * time.Second)},
		"oldest.txt":   &fstest.MapFile{ModTime: now.Add(-3 * time.Second)},
		"middle.txt":   &fstest.MapFile{ModTime: now.Add(-2 * time.Second)},
	}
	files, err := recent.Files(fsys, 10, "")
	if err != nil {
		t.Fatal(err)
	}
	if files[0].Path != "oldest.txt" {
		t.Errorf("oldest.txt is not sorted first")
	}
	if files[1].Path != "middle.txt" {
		t.Errorf("middle.txt is not sorted in the middle")
	}
	if files[2].Path != "youngest.txt" {
		t.Errorf("youngest.txt is not sorted last")
	}
}

func TestFilesExcludesPaths(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"file1.txt":                {},
		".git/file2.txt":           {},
		"file3.txt":                {},
		"tmp/.terraform/file4.txt": {},
	}
	files, err := recent.Files(fsys, 3, `\.git|\.terraform`)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("wanted %d, got %d files", 2, len(files))
	}
}
