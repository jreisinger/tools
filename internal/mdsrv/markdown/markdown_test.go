package markdown

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestToHTML(t *testing.T) {
	md := []byte(`# Title`)
	h, err := toHTML(md)
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}
	got := string(h)
	want := "<h1 id=\"table-of-contents\">Table of Contents</h1>\n<ul>\n<li>\nTitle</li>\n</ul>\n<h1>Title</h1>\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFiles(t *testing.T) {
	tests := []struct {
		fsys fs.FS
		want []string
	}{
		{fstest.MapFS{}, []string{}},
		{fstest.MapFS{"dir.md": &fstest.MapFile{Mode: fs.ModeDir}}, []string{}},
		{fstest.MapFS{
			"dir/file.md": {}, "dir/subdir/file.md": {}, "file.go": {}},
			[]string{"dir/file.md", "dir/subdir/file.md"}},
	}
	for i, test := range tests {
		got, err := Files(test.fsys)
		if err != nil {
			t.Fatalf("Files failed: %v", err)
		}
		if !eq(got, test.want) {
			t.Errorf("test %d: got %v, want %v", i, got, test.want)
		}
	}
}

func TestChangeExt(t *testing.T) {
	tests := []struct {
		path string
		ext  string
		want string
	}{
		{"file.md", ".html", "file.html"},
		{"file.txt", ".html", "file.txt"},
		{"dir/subdir/FILE.md", ".html", "dir/subdir/FILE.html"},
	}
	for _, test := range tests {
		got := ChangeExt(test.path, test.ext)
		if got != test.want {
			t.Errorf("got %q, want %q", got, test.want)
		}
	}
}

// --- test helpers ---

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
