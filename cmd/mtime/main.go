// Mtime recursively finds all files in a directory (defaults to .) and prints
// them sorted by modification time.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func main() {
	var dir string
	if len(os.Args) == 1 {
		dir = "."
	} else {
		dir = os.Args[1]
	}

	files := find(dir)
	sort.Sort(sort.Reverse(byModtime(files)))
	printFiles(files)
}

type file struct {
	path    string
	modtime time.Time
}

func find(dir string) []file {
	var files []file
	visit := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "mtime: %v\n", err)
			return filepath.SkipDir
		}
		if !entry.IsDir() {
			fi, err := entry.Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "mtime: %v\n", err)
			} else {
				files = append(files, file{
					path:    path,
					modtime: fi.ModTime(),
				})
			}
		}
		return nil
	}
	filepath.WalkDir(dir, visit)
	return files
}

type byModtime []file

func (x byModtime) Len() int           { return len(x) }
func (x byModtime) Less(i, j int) bool { return x[i].modtime.After(x[j].modtime) }
func (x byModtime) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func printFiles(files []file) {
	for _, f := range files {
		fmt.Printf("%-4.0f days ago\t%s\n", days(f.modtime), f.path)
	}
}

func days(t time.Time) float64 {
	age := time.Since(t)
	return age.Hours() / 24.0
}
