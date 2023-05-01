package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

type ArboristConf struct {
	OrgName        string   `yaml:"org_name"`
	RepoName       string   `yaml:"repo_name"`
	ExcludePattern []string `yaml:"exclude_pattern"`
}

type GHRepo struct {
	Org           string
	Name          string
	DefaultBranch string
	Branches      []GHBranch
}

type GHBranch struct {
	BranchName string
	AheadBy    int
	BehindBy   int
}

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

func parse_conf_file() ArboristConf {
	raw_conf, err := ioutil.ReadFile("arborist.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var conf ArboristConf
	err = yaml.Unmarshal(raw_conf, &conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}

func get_default_branch(ctx context.Context, client *github.Client, repo GHRepo) string {
	r, _, err := client.Repositories.Get(ctx, repo.Org, repo.Name)
	if err != nil {
		log.Fatal(err)
	}
	return r.GetDefaultBranch()
}

func main() {
	gh_token := os.Getenv("GITHUB_TOKEN")
	if gh_token == "" {
		log.Fatal("GITHUB_TOKEN env var is not defined")
	}

	conf := parse_conf_file()
	repo := GHRepo{
		Org:  conf.OrgName,
		Name: conf.RepoName,
	}

	ctx := context.Background()
	client := gh_client(ctx, gh_token)
	repo.DefaultBranch = get_default_branch(ctx, client, repo)
	fmt.Println("default branch is:", repo.DefaultBranch)

	branches, _, err := client.Repositories.ListBranches(ctx, repo.Org, repo.Name, nil)
	if err != nil {
		log.Fatal(err)
	}

	var notta_branches []string
BRANCH:
	for _, b := range branches {
		branch_name := *b.Name

		// skip comparing the default branch against itself
		if branch_name == repo.DefaultBranch {
			continue
		}

		// compare branch against the default branch
		compare, _, err := client.Repositories.CompareCommits(ctx, repo.Org, repo.Name, repo.DefaultBranch, branch_name, nil)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(branch_name, "--", "ahead by:", compare.GetAheadBy(), "behind by", compare.GetBehindBy())

		exclude := conf.ExcludePattern
		for _, pattern := range exclude {
			match, err := regexp.MatchString(pattern, branch_name)
			if err != nil {
				log.Fatal(err)
			}
			if match == true {
				fmt.Println("ignoring branch:", branch_name, "because it matched exclude_pattern:", pattern)
				continue BRANCH
			}
		}

		if compare.GetAheadBy() == 0 {
			notta_branches = append(notta_branches, branch_name)
		}
	}

	fmt.Println("branches with are not ahead")
	for _, b := range notta_branches {
		fmt.Println(b)
	}
}
