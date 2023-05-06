// Package ghfind searches files in a GitHub repository.
package ghfind

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// GitHub API base URL.
const Base = "https://api.github.com"

// File represents a file in a GitHub repository.
type File struct {
	Path string
	// Type   string // "tree" means directory, "blob" means file
	repo   string
	branch string
}

// Files returns all the files in a repo. Repo is in the <owner>/<repo> format.
func Files(repo, branch string) ([]File, error) {
	url := fmt.Sprintf("%s/repos/%s/git/trees/%s?recursive=1", Base, repo, branch)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	files, err := unmarshalFiles(b)
	if err != nil {
		return nil, err
	}

	// Inject repo and branch so it can be used by Print.
	for i := range files {
		files[i].repo = repo
		files[i].branch = branch
	}

	return files, nil
}

func unmarshalFiles(data []byte) ([]File, error) {
	var jsonResponse struct{ Tree []File }
	if err := json.Unmarshal(data, &jsonResponse); err != nil {
		return nil, err
	}
	return jsonResponse.Tree, nil
}

// Print prints files that have a path matching the regex. If regex is empty,
// all files are printed. Files are printed as URL links that you can click.
func Print(files []File, regex string) {
	for _, file := range files {
		prefix := fmt.Sprintf("https://github.com/%s/blob/%s/", file.repo, file.branch)
		if regex != "" {
			pathRe := regexp.MustCompile("(?i)" + regex)
			if pathRe.MatchString(file.Path) {
				fmt.Println(prefix + file.Path)
			}
		} else {
			fmt.Println(prefix + file.Path)
		}
	}
}
