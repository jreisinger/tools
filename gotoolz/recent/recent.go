package recent

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sort"
	"time"
)

type File struct {
	Path    string
	ModTime time.Time
}

func Files(fsys fs.FS, n int, excludePath string) ([]File, error) {
	excludePathRE, err := parseRegexp(excludePath)
	if err != nil {
		return nil, err
	}
	var files []File
	err = fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, e error) error {
		if excludePathRE != nil && excludePathRE.MatchString(p) {
			return nil
		}
		if e != nil {
			if errors.Is(e, fs.ErrPermission) {
				fmt.Fprintf(os.Stderr, "recent: %v\n", e)
			} else {
				return e
			}
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		files = append(files, File{Path: p, ModTime: info.ModTime()})
		return nil
	})
	sortByModTime(files)
	return getLastNFiles(files, n), err
}

func parseRegexp(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, nil
	}
	return regexp.Compile(pattern)
}

func getLastNFiles(files []File, n int) []File {
	if n > len(files) {
		n = len(files)
	}
	return files[len(files)-n:]
}

func sortByModTime(files []File) {
	oldestToYoungest := func(i, j int) bool {
		return files[i].ModTime.Before(files[j].ModTime)
	}
	sort.Slice(files, oldestToYoungest)
}
