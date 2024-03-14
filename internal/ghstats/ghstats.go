package ghstats

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

type Stat struct {
	Repository       string
	Pushed           time.Time
	Stars            int
	UniqueVisitors   int
	UniqueCloners    int
	ReleaseDownloads int
}

type Stats []Stat

func Get(user, token string) (Stats, error) {
	ctx := context.Background()
	client := getClient(ctx, token)

	repos, err := getAllRepos(ctx, client, user)
	if err != nil {
		return nil, err
	}

	ch := make(chan Stat)

	for _, r := range repos {
		go func(r *github.Repository) {
			views, _, err := client.Repositories.ListTrafficViews(ctx, user, r.GetName(), nil)
			if err != nil {
				log.Print(err)
			}

			clones, _, err := client.Repositories.ListTrafficClones(ctx, user, *r.Name, nil)
			if err != nil {
				log.Print(err)
			}

			var releaseDownloads int
			releases, _, err := client.Repositories.ListReleases(ctx, user, r.GetName(), nil)
			if err != nil {
				log.Print(err)
			}
			for _, r := range releases {
				for _, a := range r.Assets {
					releaseDownloads += *a.DownloadCount
				}
			}

			stat := Stat{
				Repository:       r.GetName(),
				Pushed:           r.PushedAt.Time,
				Stars:            r.GetStargazersCount(),
				UniqueVisitors:   views.GetUniques(),
				UniqueCloners:    clones.GetUniques(),
				ReleaseDownloads: releaseDownloads,
			}
			ch <- stat
		}(r)
	}

	var stats Stats
	for range repos {
		stats = append(stats, <-ch)
	}

	return stats, nil
}

func getClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	hc := oauth2.NewClient(ctx, ts)
	return github.NewClient(hc)
}

// getAllRepos returns all user's repositories. Set user to empty string for
// repositories of the authenticated user.
func getAllRepos(ctx context.Context, ghc *github.Client, user string) ([]*github.Repository, error) {
	opt := &github.RepositoryListOptions{ListOptions: github.ListOptions{PerPage: 10}}
	var allRepos []*github.Repository
	for {
		repos, resp, err := ghc.Repositories.List(ctx, user, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos, nil
}

func (stats Stats) Sort(column int) {
	sort.Sort(customSort{stats: stats, less: func(x, y Stat) bool {
		switch column {
		case 1:
			if x.Repository != y.Repository {
				return x.Repository < y.Repository
			}
		case 2:
			if x.Pushed != y.Pushed {
				return x.Pushed.After(y.Pushed)
			}
		case 3:
			if x.Stars != y.Stars {
				return x.Stars > y.Stars
			}
		case 4:
			if x.UniqueVisitors != y.UniqueVisitors {
				return x.UniqueVisitors > y.UniqueVisitors
			}
		case 5:
			if x.UniqueCloners != y.UniqueCloners {
				return x.UniqueCloners > y.UniqueCloners
			}
		case 6:
			if x.ReleaseDownloads != y.ReleaseDownloads {
				return x.ReleaseDownloads > y.ReleaseDownloads
			}
		default:
			log.Fatalf("can't sort by column %d", column)
		}
		if x.Repository != y.Repository {
			return x.Repository < y.Repository
		}
		return false
	}})
}

type customSort struct {
	stats []Stat
	less  func(x, y Stat) bool
}

func (x customSort) Len() int           { return len(x.stats) }
func (x customSort) Less(i, j int) bool { return x.less(x.stats[i], x.stats[j]) }
func (x customSort) Swap(i, j int)      { x.stats[i], x.stats[j] = x.stats[j], x.stats[i] }

func (stats Stats) Print(topN int) {
	const format = "%v\t%v\t%v\t%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Repository", "Pushed", "Stars", "Visitors (2w)", "Cloners (2w)", "Release downloads")
	fmt.Fprintf(tw, format, "----------", "------", "-----", "-------------", "------------", "-----------------")
	var n int
	var total struct {
		stars     int
		visitors  int
		cloners   int
		downloads int
	}
	for _, s := range stats {
		if n == topN {
			break
		}
		n++
		total.stars += s.Stars
		total.visitors += s.UniqueVisitors
		total.cloners += s.UniqueCloners
		total.downloads += s.ReleaseDownloads
		fmt.Fprintf(tw, format, s.Repository, s.Pushed.Format("2006-01-02"), s.Stars, s.UniqueVisitors, s.UniqueCloners, s.ReleaseDownloads)
	}

	// Print footer.
	fmt.Fprintf(tw, format, "          ", "      ", "-----", "-------------", "------------", "-----------------")
	fmt.Fprintf(tw, format, "", "", total.stars, total.visitors, total.cloners, total.downloads)

	tw.Flush()
}
