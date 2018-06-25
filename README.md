# gh-contrib-stats

*TODO: think of a better name..*

Fetches contributor stats for a given date range.

**NB:** The GitHub API returns contributor stats grouped by week beginning. This should be considered when passing a date range.


## Install
`go get github.com/luke-davies/gh-contrib-stats`

## GitHub 202
When first querying a repo you are likely to get:

`[ListContributorStats] [GitHub Error] GitHub sent a 202, meaning they don't have those stats ready. Try again in a minute`

This means what it says. Try again in a minute. GitHub needs time to calculate the contributor stats.

## Examples
```
# get all contributors stats for golang/go:
gh-contrib-stats golang/go

# get all contributors stats from a given date to today:
gh-contrib-stats --from 2018-05-10 golang/go

# get all contributors stats up to a given date:
gh-contrib-stats --to 2018-06-14 golang/go

# get all contributors stats between two dates:
gh-contrib-stats --from 2018-05-10 --to 2018-06-14 golang/go

# get all contributors stats in last N weeks:
gh-contrib-stats --weeks 4 golang/go

# get all contributors stats in last N months:
gh-contrib-stats --months 2 golang/go

# get all contributors stats in last N years:
gh-contrib-stats --years 1 golang/go

# pass --all to include contributors that have 0 commits in the given date range:
gh-contrib-stats --all --from 2018-05-10 golang/go
gh-contrib-stats --all --to 2018-06-14 golang/go
gh-contrib-stats --all --from 2018-05-10 --to 2018-06-14 golang/go
gh-contrib-stats --all --weeks 4 golang/go
gh-contrib-stats --all --months 2 golang/go
gh-contrib-stats --all --years 1 golang/go

```

## Run Tests

`cd $GOPATH/src/github.com/luke-davies/gh-contrib-stats`:

`go test ./...`
