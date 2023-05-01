package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	gh_token := os.Getenv("GITHUB_TOKEN")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gh_token},
	)
	tc := oauth2.NewClient(ctx, ts)
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)
	if err != nil {
		panic(err)
	}

	client := github.NewClient(rateLimiter)

	orgs, _, err := client.Organizations.List(context.Background(), "jhoblitt", nil)

	if err != nil {
		fmt.Println("that sucks: ", err)
	}

	fmt.Println(orgs)
}
