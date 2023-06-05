# Arborist

![A Tidy Repo Forest](./images/4193380335_A_lush_forest_of_tall_trees_with_branches_being_cu_xl-beta-v2-2-2.png)

Prune orphaned branches from a forest of repos.

## Synopsys

**Arborist** is intended to prevent orphaned or dead branches from accumulating
in multi-repo projects. It compares branches across any number of git repos
hosted on GitHub, although it may also be used to prune branches from a single
repo.

Branches with commits "ahead" of a repo's "default branch" are always
preserved.  If a branch is even with or behind the "default branch", all other
repos are inspected for a branch of the same name which is ahead of master. The
branch will be preserved if any other repo's matching branch is ahead of its
"default branch", otherwise it is a candidate for removal.

## Installation

### Binary Releases

Binaries are published as [GitHub
Releases](https://github.com/jhoblitt/arborist/releases).

### Downloading Dangerous Shellcode From the Internet And Executing It

```bash
mkdir -p ~/bin
cd ~/bin
\curl -sSL https://raw.githubusercontent.com/jhoblitt/arborist/main/install.sh | bash -s
```

## Usage

```bash
$ GITHUB_TOKEN=<token> arborist
```

### GitHub API Token

`arborist` uses GitHub's API and requires an API token to function. This may be
provided by setting the env var `GITHUB_TOKEN` or specified with the
`--github-token` flag. The env var should generally be preferred as this does
not result in a "secret" being visible in the host's process table.

### Config File

Most configuration is in the form of a `.arborist.yaml` file. By default, this
file is expected to be in the same path from which `arborist` was invoked.  The
location of the configuration file may be overridden with the `--conf` flag.

```bash
GITHUB_TOKEN=... arborist --conf /foo/bar/baz.yaml
```

The config file is in vanilla [`yaml`](https://yaml.org/) format.

Example `.arborist.yaml`:

```
---
noop: true  # default
repos:
  - repo: jhoblitt/arborist-test1
    noop: false
  - repo: jhoblitt/arborist-test2
    noop: true  # default
exclude_patterns:
  - main
```

#### Top Level Config

| Key             | Description|
| -----           | -----------------------|
| noop            | A "master arm switch" for branch deletion.  When `true` no branches may be deleted.  When `false`, branches in a repo may be deleted but only if the `noop: false` is also set on the repo directly. Defaults to `true`. |
| repos           | A Sequence (list) of Mappings (hashes) that configure which repo(s) make up the forest in [Repo Config](#repo-config) format. At least one repo must be specified. |
| exclude_pattern | A Sequence (list) of branch name patterns to ignore. Patterns are [Golang style regular expressions](https://github.com/google/re2/wiki/Syntax). Optional. |

#### Repo Config

| Key   | Description                                                      |
| ----- | -----------------------                                          |
| repo  | Name of GitHub repo in `<org>/<name>` format. Required.          |
| noop  | Branches may be deleted when set to `false`. Defaults to `true`. |

## OCI Image

OCI images are published to [ghcr.io/jhoblitt/arborist](https://github.com/jhoblitt/arborist/pkgs/container/arborist).

### Example `docker run`

```bash
$ docker run -ti -e GITHUB_TOKEN=$GITHUB_TOKEN -v $(pwd):$(pwd) -w $(pwd) ghcr.io/jhoblitt/arborist:latest
```

## GitHub Action

A GitHub Action is available to help automatically maintain a healthy forest.
See: [`jhoblitt/arborist-action`](https://github.com/jhoblitt/arborist-action)
