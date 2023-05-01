package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
)

func gh_client(ctx context.Context, gh_token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gh_token},
	)
	tc := oauth2.NewClient(ctx, ts)
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)
	if err != nil {
		log.Fatal(err)
	}

	return github.NewClient(rateLimiter)
}

func main() {
	org_name := "lsst-it"
	repo_name := "lsst-control"
	gh_token := os.Getenv("GITHUB_TOKEN")

	ctx := context.Background()
	client := gh_client(ctx, gh_token)

	repo, _, err := client.Repositories.Get(ctx, org_name, repo_name)
	if err != nil {
		log.Fatal(err)
	}
	default_branch := repo.GetDefaultBranch()
	fmt.Println("default branch is:", default_branch)

	branches, _, err := client.Repositories.ListBranches(ctx, org_name, repo_name, nil)
	if err != nil {
		log.Fatal(err)
	}

	var notta_branches []string
	for _, b := range branches {
		name := *b.Name

		// skip comparing the default branch against itself
		if *b.Name == default_branch {
			continue
		}

		// compare branch against the default branch
		compare, _, err := client.Repositories.CompareCommits(ctx, org_name, repo_name, default_branch, name, nil)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(*b.Name, "--", "ahead by:", compare.GetAheadBy(), "behind by", compare.GetBehindBy())

		match, err := regexp.MatchString("production", name)
		if match == true {
			continue
		}

		if compare.GetAheadBy() == 0 {
			notta_branches = append(notta_branches, name)
		}
	}

	fmt.Println("branches with are not ahead")
	for _, b := range notta_branches {
		fmt.Println(b)
	}
}
