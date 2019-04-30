package main

import (
	"context"
	"flag"
	"os"
	"path"
	"time"

	"github.com/yagi5/gh-auto-merge/github"
)

var (
	interval    time.Duration
	mergeMethod string
	prs         []string
)

func main() {
	parseOpt()

	ctx := context.Background()

	ghToken := os.Getenv("GITHUB_TOKEN")
	gh := github.New(ghToken, mergeMethod, interval, repoPath())

	go func() {
		gh.AutoMerge(ctx)
	}()

	for _, pr := range prs {
		github.PushPR(ctx, pr)
	}

	<-ctx.Done()
}

func parseOpt() {
	var (
		merge  bool
		squash bool
		rebase bool
	)
	flag.DurationVar(&interval, "interval", 10*time.Second, "interval for merge attempt retry")
	flag.DurationVar(&interval, "i", 10*time.Second, "interval for merge attempt retry(short)")

	flag.BoolVar(&merge, "merge", true, "use 'merge' as merge method. (default)")
	flag.BoolVar(&squash, "squash", false, "use 'squash' as merge method")
	flag.BoolVar(&rebase, "rebase", false, "use 'rebase' as merge method")

	flag.Parse()

	prs = flag.Args()

	if squash {
		mergeMethod = "squash"
		return
	}
	if rebase {
		mergeMethod = "rebase"
		return
	}
	mergeMethod = "merge" // default
}

func repoPath() string {
	gopath := os.Getenv("GOPATH")
	return path.Join(gopath, "src", "github.com")
}
