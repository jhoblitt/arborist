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
	Branches      map[string]GHBranch
}

type GHBranch struct {
	Name     string
	AheadBy  int
	BehindBy int
}

func NewGHRepo(ctx context.Context, client *github.Client, org, name string) GHRepo {
	r := GHRepo{
		Org:      org,
		Name:     name,
		Branches: make(map[string]GHBranch),
	}

	r.DefaultBranch = get_default_branch(ctx, client, r)
	fmt.Printf("%s default branch is: %s\n", r.FullName(), r.DefaultBranch)

	branches, _, err := client.Repositories.ListBranches(ctx, r.Org, r.Name, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, b := range branches {
		branch_name := *b.Name

		// skip comparing the default branch against itself
		if branch_name == r.DefaultBranch {
			continue
		}

		// compare branch against the default branch
		compare, _, err := client.Repositories.CompareCommits(ctx, r.Org, r.Name, r.DefaultBranch, branch_name, nil)
		if err != nil {
			log.Fatal(err)
		}

		r.Branches[branch_name] = GHBranch{
			Name:     branch_name,
			AheadBy:  compare.GetAheadBy(),
			BehindBy: compare.GetAheadBy(),
		}
	}

	return r
}

func (r GHRepo) FullName() string {
	return fmt.Sprintf("%s/%s", r.Org, r.Name)
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

// https://stackoverflow.com/a/57213476
func RemoveIndex(s []GHRepo, index int) []GHRepo {
	ret := make([]GHRepo, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
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

	safe_branches := map[string]GHRepo{}

	for i, r := range project_repos {
		other_repos := RemoveIndex(project_repos, i)

	BRANCH:
		for _, b := range r.Branches {
			fmt.Printf("%s/%s:", r.Org, r.Name)
			fmt.Println(b.Name, "--", "ahead by:", b.AheadBy, "behind by", b.BehindBy)

			// Check to see if this branch is already known to be ahead.
			safer, ok := safe_branches[b.Name]
			if ok {
				fmt.Println("ignoring branch:", b.Name, "because it is known to be ahead in", safer.FullName())
				continue BRANCH
			}

			// if this branch is ahead of the default branch, we want to preserve it
			// across all repos, even if it not ahead of the default branch in other
			// repo(s).
			if b.AheadBy != 0 {
				fmt.Println("ignoring branch:", b.Name, "because it is ahead")
				safe_branches[b.Name] = r
				continue BRANCH
			}

			exclude := conf.ExcludePatterns
			for _, pattern := range exclude {
				match, err := regexp.MatchString(pattern, b.Name)
				if err != nil {
					log.Fatal(err)
				}
				if match == true {
					// This branch isn't being added to safe_branches so we can report if
					// a branch is ignored because it is ahead in a different repo or it
					// has matched an exclude_pattern.
					fmt.Println("ignoring branch:", b.Name, "because it matched exclude_pattern:", pattern)
					continue BRANCH
				}
			}

			// Check for this branch in other repos. If any repo has this branch
			// ahead of the default branch, then this branch name is considered safe
			// across all repos and there's no need to keep checking any additional
			// repos.
			for _, otherr := range other_repos {
				otherb, ok := otherr.Branches[b.Name]

				if ok && otherb.AheadBy != 0 {
					fmt.Println("ignoring branch:", b.Name, "because it is ahead in", otherr.FullName())
					safe_branches[b.Name] = r
					continue BRANCH
				}
			}
		}
	}
}
