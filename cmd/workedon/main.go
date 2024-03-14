// Workedon tells you what you (or others) have worked on. It gets this
// information from git commit logs.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type directory struct {
	path    string
	changes int
	authors []string
	repo    *git.Repository
	files   []file
}

type file struct {
	path    string
	changes int
	authors []string
}

var (
	author = flag.String("author", "", "only changes by `this` author")
	days   = flag.Int("days", 7, "changes made in last `n` days")
	files  = flag.Bool("files", false, "changes per file (default is per repo)")
	ignore = flag.String("ignore", "", "ignore `filename` (e.g. LICENSE)")
	pull   = flag.Bool("pull", false, "pull the repo before parsing its logs")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix(os.Args[0] + ": ")

	flag.Usage = func() {
		desc := "What git-tracked stuff have you (or others) worked on."
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\n%s [flags] repo [repo ...]\n", desc, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nEXAMPLE\n  workedon ~/github.com/*/*\n")
	}

	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	in := make(chan directory)
	out := make(chan directory)

	var wg sync.WaitGroup

	// Send directories containing a git repo down the in channel.
	wg.Add(1)
	go func() {
		// LIFO order!
		defer wg.Done()
		defer close(in)

		for _, path := range flag.Args() {
			repo, err := git.PlainOpen(path)
			if err != nil {
				log.Printf("%s: %v", path, err)
				continue
			}

			in <- directory{
				path: path,
				repo: repo,
			}
		}
	}()

	// Get directories from the in channel, enrich them with info from
	// parsed repo logs and send them down the out channel.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dir := range in {
				since := time.Hour * 24 * time.Duration(*days)
				files, err := parseRepoLogs(dir.repo, pull, author, &since)
				if err != nil {
					switch err.(type) {
					case *pullError:
						log.Printf("pulling repo %s: %v", dir.path, err)
					default:
						log.Fatalf("parsing repo %s: %v", dir.path, err)
					}
				}
				for _, f := range files {
					if *ignore != "" && filepath.Base(f.path) == *ignore {
						continue
					}
					dir.changes += f.changes
					dir.authors = append(dir.authors, f.authors...)
					dir.files = append(dir.files, f)
				}
				out <- dir
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	reportResults(out)
}

func reportResults(out chan directory) {
	var totalChanges int
	var directories []directory
	for dir := range out {
		if len(dir.files) == 0 {
			continue
		}
		totalChanges += dir.changes
		directories = append(directories, dir)
	}

	if len(directories) == 0 {
		return
	}

	const format = "%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "PATH", "CHANGES", "AUTHORS")

	sort.Sort(sort.Reverse(byDirChanges(directories)))
	for _, dir := range directories {
		if *files {
			sort.Sort(sort.Reverse(byFileChanges(dir.files)))
			for _, f := range dir.files {
				changes := fmt.Sprintf("%2.0f%% (%d)", float64(f.changes)/float64(totalChanges)*100, f.changes)
				authors := strings.Join(uniq(f.authors), ", ")
				fmt.Fprintf(tw, format, filepath.Join(dir.path, f.path), changes, authors)
			}
		} else {
			changes := fmt.Sprintf("%2.0f%% (%d)", float64(dir.changes)/float64(totalChanges)*100, dir.changes)
			authors := strings.Join(uniq(dir.authors), ", ")
			fmt.Fprintf(tw, format, dir.path, changes, authors)
		}
	}

	tw.Flush()
}

type byFileChanges []file

func (x byFileChanges) Len() int           { return len(x) }
func (x byFileChanges) Less(i, j int) bool { return x[i].changes < x[j].changes }
func (x byFileChanges) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type byDirChanges []directory

func (x byDirChanges) Len() int           { return len(x) }
func (x byDirChanges) Less(i, j int) bool { return x[i].changes < x[j].changes }
func (x byDirChanges) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type pullError struct {
	Err error
}

func (e *pullError) Error() string {
	return fmt.Sprint(e.Err)
}

func parseRepoLogs(repo *git.Repository, pull *bool, author *string, since *time.Duration) (files []file, err error) {
	if *pull {
		if err := pullRepo(repo); err != nil {
			return nil, &pullError{Err: err}
		}
	}

	t := time.Now().Add(-*since)
	cIter, err := repo.Log(&git.LogOptions{Since: &t})
	if err != nil {
		return nil, err
	}

	changesPerFile := make(map[string]int)
	authorsPerFile := make(map[string][]string)
	msgsPerFile := make(map[string][]string)
	err = cIter.ForEach(func(commit *object.Commit) error {
		if *author != "" && commit.Author.Name != *author {
			return nil
		}

		stats, err := commit.Stats()
		if err != nil {
			return err
		}

		for _, stat := range stats {
			file, nChanges := parseStat(stat)
			if file != "" { // only content changes
				changesPerFile[file] += nChanges
			}

			authorsPerFile[file] = append(authorsPerFile[file], commit.Author.Name)

			lines := strings.Split(commit.Message, "\n")
			msgsPerFile[file] = append(msgsPerFile[file], lines[0])
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	for f, c := range changesPerFile {
		files = append(files, file{
			path:    f,
			changes: c,
			authors: uniq(authorsPerFile[f]),
		})
	}

	return
}

func pullRepo(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	privateKeyFile := filepath.Join(home, ".ssh", "id_rsa")

	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{
		Auth: publicKeys,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}
	return nil
}

func uniq(ss []string) []string {
	keys := make(map[string]bool)
	uniq := []string{}
	for _, s := range ss {
		if _, ok := keys[s]; !ok {
			keys[s] = true
			uniq = append(uniq, s)
		}
	}
	return uniq
}

func parseStat(stat object.FileStat) (file string, nChanges int) {
	count := make(map[string]int)
	if _, ok := count[stat.Name]; !ok {
		count[stat.Name]++
	}
	file = stat.Name
	nChanges += stat.Addition
	nChanges += stat.Deletion
	for _, v := range count {
		if v > 1 {
			log.Fatalf("didn't expect this: %v", count)
		}
	}
	return
}
