package github

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
)

// PullRequest is pull request
type pullRequest struct {
	s      string
	owner  string
	repo   string
	number int
	pr     *github.PullRequest
}

// NewPR returns PR object
// also checks given url as valid as GitHub Pull request URL format
func (gh *GitHub) newPR(ctx context.Context, s string) (*pullRequest, error) {
	owner, repo, number, err := gh.parsePRURL(s)
	if err != nil {
		return nil, err
	}
	pr := &pullRequest{s: s, owner: owner, repo: repo, number: number}
	p, _, err := gh.client.PullRequests.Get(ctx, pr.owner, pr.repo, pr.number)
	if err != nil {
		return nil, err
	}
	pr.pr = p
	return pr, nil
}

// ParsePRURL ...
func (gh *GitHub) parsePRURL(s string) (string, string, int, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", "", 0, err
	}
	// Pull request url format: https://github.com/yagi5/gh-auto-merge/pull/1
	matches, err := path.Match("/*/*/pull/*", u.Path)
	if err != nil {
		return "", "", 0, err
	}
	if !matches {
		return "", "", 0, fmt.Errorf("invalid format as pull request URL")
	}
	splitted := strings.Split(u.Path, "/") // after Split: ["", "yagi5", "gh-auto-merge", "pull", "1"] because of head slash
	number, err := strconv.Atoi(splitted[4])
	if err != nil {
		return "", "", 0, err
	}
	return splitted[1], splitted[2], number, nil
}

func (pr *pullRequest) mergeable() bool {
	return pr.pr.GetMergeable()
}

func (pr *pullRequest) merged() bool {
	return pr.pr.GetMerged()
}

func (pr *pullRequest) closed() bool {
	return !pr.pr.GetClosedAt().IsZero()
}

func (pr *pullRequest) title() string {
	return pr.pr.GetTitle()
}

func (pr *pullRequest) headBranch() string {
	return pr.pr.GetBase().GetLabel()
}
