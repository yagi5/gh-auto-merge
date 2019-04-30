package github

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	queue = make(chan string, 1)

	errConflict         = errors.New("errConflict")
	errStillDraft       = errors.New("errStillDraft")
	errReviewIsRequired = errors.New("errReviewIsRequired")
	errNotUpToDate      = errors.New("errNotUpToDate")
)

// GitHub is GitHub client
type GitHub struct {
	client *github.Client
}

// New return GitHub client
func New(token string) *GitHub {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	gh := github.NewClient(tc)
	return &GitHub{client: gh}
}

// PushPR pushes waiting-merge PR into queue
func PushPR(ctx context.Context, pullRequestURL string) {
	queue <- pullRequestURL
}

// AutoMerge merges given PR automatically
func (gh *GitHub) AutoMerge(ctx context.Context) {
	for {
		p := <-queue
		time.Sleep(10 * time.Second) // backoff

		pr, err := gh.newPR(ctx, p)
		if err != nil {
			log.Printf("invalid PR url: %s, err: %s\n", p, err)
			continue
		}

		if pr.merged() {
			log.Printf("PR is alread merged: %s\n", p)
			continue
		}

		if pr.closed() {
			log.Printf("PR is alread closed: %s\n", p)
			continue
		}

		if !pr.mergeable() {
			log.Printf("PR has conflicts. Fix it first: %s\n", p)
			continue
		}

		log.Printf("try merge: %s\n", p)

		err = gh.merge(ctx, pr)
		if err == nil {
			log.Printf("merge succeeded: %s\n", p)
			continue
		}

		if err == errStillDraft {
			log.Printf("PR is still draft. Try again later: %s\n", p)
			queue <- p
			continue
		}

		if err == errConflict {
			log.Printf("PR has conflicts. Fix it first: %s\n", p)
			queue <- p
			continue
		}

		if err == errReviewIsRequired {
			log.Printf("PR is not reviewed yet: %s\n", p)
			queue <- p
			continue
		}

		// Coming here, pr is needed to be up-to-date
		log.Printf("PR is not up-to-date. Try to update: %s\n", p)
		queue <- p
	}
}

func (gh *GitHub) merge(ctx context.Context, p *pullRequest) error {
	// https://developer.github.com/v3/pulls/#merge-a-pull-request-merge-button
	_, _, err := gh.client.PullRequests.Merge(ctx, p.owner, p.repo, p.number, p.title(), &github.PullRequestOptions{})
	rerr := err.(*github.ErrorResponse)
	code := rerr.Response.StatusCode
	switch code {
	case http.StatusConflict:
		return errConflict
	case http.StatusMethodNotAllowed:
		return messageToErr(rerr.Message)
	default:
		return nil // 200
	}
}

func messageToErr(msg string) error {
	if strings.Contains(msg, "still a draft") {
		return errStillDraft
	}
	if strings.Contains(msg, "review is required") {
		return errReviewIsRequired
	}
	// probably need to cover `PR is already approved but some checks are failing`
	return errNotUpToDate
}
