# RepoScrapper

## Overview

## Important commands

Note: When run from the `src/` directory

### Updating Repositories

`run go . repos update`

Used to pull down latest repos, update existing, or delete removed repos. |

### Scan with Conftest

Note: Newest command and still little unstable, best to run in debugger to help with troubleshooting.

Currently output is hardcoded to `output.csv`

`run go . conftest` - Simplest command, this will run against all repositories known to this app and scan all supported files against all rules.

```
Flags:
  -f, --file-pattern string          regex for selected filenames (default "^(dockerfile|(.*(\\.dockerfile|\\.yml|\\.yaml|\\.tf|\\.tfvars|\\.json)))$")
  -h, --help                         help for conftest
  -i, --ignore-pattern string        regex for filepaths to ignore (default "(\\/allure-results\\/)|(\\/mock-data\\/)|(\\/.{0,3}test.{0,3}\\/)|(\\/node_modules\\/)")
  -n, --namespaces stringArray       namespaces to use
  -p, --policy-location string       location of conftest policies (default "~/code/conftest-rego-policies/policies/nonblocking")
  -r, --repositories stringArray     repositories to scan
  -l, --repository-location string   base location of repositories to scan (default "~/.reposcrapper/repos")
```

## Dev Setup

This application currently make several assumptions when it comes to the developers local environment setup.

### API Tokens

It assumes that the following API tokens are environment variables with the given name

| Service | Env Key |
| --------| --------|
| GitHub | GITHUB_TOKEN |
| Snyk | SNYK_TOKEN |

### Important Directories and Files

I can't remember if it will create the `~/.reposcrapper` and `~/.reposcrapper/repos` directories or if you as the developer has too... TBD

| Directory/File | Notes |
| -------------- | ----- |
| `~/.reposcrapper/` | Main directory where most long term information |
| `~/.reposcrapper/repos/` | Directory where cloned repositories will be stored |
| `~/.reposcrapper/cache/ ` | I am not sure if this is used anymore, might have been replaced by repos... TBD |
| `~/.reposcrapper/db.raw.json` | Data store for all the repositories and metadata around them. |
| `~/.reposcrapper/filelist.txt` | A list of each file in the repos folder along with its size.  Used for simple grepping to find what files are around or looking for size of large files |

