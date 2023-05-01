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
	Repos           []RepoConf `yaml:"repos"`
	ExcludePatterns []string   `yaml:"exclude_patterns"`
}

type RepoConf struct {
	OrgName  string `yaml:"org_name"`
	RepoName string `yaml:"repo_name"`
}

type GHRepo struct {
	Org           string
	Name          string
	DefaultBranch string
	Branches      []GHBranch
}

type GHBranch struct {
	Name     string
	AheadBy  int
	BehindBy int
}

func NewGHRepo(ctx context.Context, client *github.Client, org, name string) GHRepo {
	repo := GHRepo{
		Org:  org,
		Name: name,
	}

	repo.DefaultBranch = get_default_branch(ctx, client, repo)
	fmt.Printf("%s/%s default branch is: %s\n", repo.Org, repo.Name, repo.DefaultBranch)

	branches, _, err := client.Repositories.ListBranches(ctx, repo.Org, repo.Name, nil)
	if err != nil {
		log.Fatal(err)
	}

	var branch_info []GHBranch
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

		branch_info = append(branch_info, GHBranch{
			Name:     branch_name,
			AheadBy:  compare.GetAheadBy(),
			BehindBy: compare.GetAheadBy(),
		})
	}

	repo.Branches = branch_info

	return repo
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
	ctx := context.Background()
	client := gh_client(ctx, gh_token)

	var project_repos []GHRepo
	for _, r := range conf.Repos {
		project_repos = append(project_repos, NewGHRepo(ctx, client, r.OrgName, r.RepoName))
	}

	for _, r := range project_repos {
		var notta_branches []string
	BRANCH:
		for _, b := range r.Branches {
			fmt.Println(b.Name, "--", "ahead by:", b.AheadBy, "behind by", b.BehindBy)

			exclude := conf.ExcludePatterns
			for _, pattern := range exclude {
				match, err := regexp.MatchString(pattern, b.Name)
				if err != nil {
					log.Fatal(err)
				}
				if match == true {
					fmt.Println("ignoring branch:", b.Name, "because it matched exclude_pattern:", pattern)
					continue BRANCH
				}
			}

			if b.AheadBy == 0 {
				notta_branches = append(notta_branches, b.Name)
			}
		}

		fmt.Println("branches with are not ahead")
		for _, b := range notta_branches {
			fmt.Println(b)
		}
	}
}
