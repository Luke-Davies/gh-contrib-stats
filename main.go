// Package main is responsible for reading input and presenting output.
// In a real application this would often be handled by a separate package.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/luke-davies/gh-contrib-stats/pkg/app"
	"github.com/luke-davies/gh-contrib-stats/pkg/github"
)

const (
	githubBaseURL = "https://api.github.com"
)

func main() {
	raw, err := parseInput()
	if err != nil {
		log.Print(err) // would log.Fatal but want usage after error
		flag.Usage()
		os.Exit(1)
	}

	inputs, err := processInput(raw)
	if err != nil {
		log.Print(err)
		flag.Usage()
		os.Exit(1)
	}

	// TODO: Move more of this into other functions to make it easier to test
	ghClient := github.Client{BaseURL: githubBaseURL}

	gcs, err := ghClient.ListContributorStats(context.Background(), inputs.Owner, inputs.Repo)
	if err != nil {
		// usage probably not helpful if they make it this far..
		log.Fatal(err.Error())
	}

	var acs []app.Contributor
	opts := app.NormaliseCalcContributionsOpts(app.CalcContrbutionsOpts{From: inputs.From, To: inputs.To})
	err = app.ValidateCalcContributionsOpts(opts)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, gc := range *gcs {
		c := app.CalcContributions(gc, opts)
		acs = append(acs, c)
	}

	if !inputs.All {
		acs = app.FilterContributors(acs, func(ac app.Contributor) bool { return ac.Stats.Commits > 0 })
	}

	printStats(acs)

}

// simplifies parseInput signature
type rawInputs struct {
	Repo   string
	From   string
	To     string
	Weeks  int
	Months int
	Years  int
	All    bool
}

// ParseInput parses flags and returns relevant 'repo-owner', `repo-name`, from`, `to` and `all`.
// - repoOwner and repoName idenitfy the GitHub repository.
// - from & to specify the date range.
// - all specifies whether to include contributors who have no contributions during the specified date range.
//
// Exits on error.
func parseInput() (rawInputs, error) {
	from := flag.String("from", "", "Lower bound (inclusive) of the date range. Format: `YYYY-MM-DD` e.g. `1966-07-30`. Can be used with --from but must be before --from. Dates on or before 0001-01-01 are ignored. Can not be used with --weeks, --months or --years.")
	to := flag.String("to", "", "Upper bound (exclusive) of the date range. Format: `YYYY-MM-DD` e.g. `1966-07-30`. Can be used with --to but must be after --from. Dates on or before 0001-01-01 are ignored. Can not be used with --weeks, --months or --years.")
	weeks := flag.Int("weeks", 0, "Set lower bound by number of weeks. Can be combined with --months and --years. Zero is ignored. Can not be used with --from and --to.")
	months := flag.Int("months", 0, "Set lower bound by number of months. Can be combined with --weeks and --years. Zero is ignored. Can not be used with --from and --to.")
	years := flag.Int("years", 0, "Set lower bound by number of years. Can be combined with --weeks and --months. Zero is ignored. Can not be used with --from and --to.")
	all := flag.Bool("all", false, "Show all contributors regardless of whether they have made contributions during the specified date range. By default, contributors without contributions in the date range are omitted.")

	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Usage of %[1]s: %[1]s [date range options] [owner]/[repo]\n\n"+
				"Retrieves contributor stats for a repository for the given date range.\n"+
				"This uses the GitHub API, which groups stats by week beginning. Therefore, stats for yesterday may not appear if the "+
				"beginning of the week is not within the date range.\n"+
				"Contributors with 0 commits in the given date range are filtered out.\n\n"+
				"Examples:\n"+
				"\t%[1]s go/golang\n"+
				"\t%[1]s --from 2017-09-01 --to 2018-02-01 go/golang\n"+
				"\t%[1]s --weeks 10 go/golang\n\n"+
				"Options for date range:\n\n",
			os.Args[0],
		)
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() > 1 {
		// because flag.Parse() cant find flags after the args...
		return rawInputs{}, errors.New("[checkFlags] only one argument expected and all flags should be specified before repository argument")
	}

	repo := flag.Arg(0)

	if repo == "" {
		return rawInputs{}, errors.New("[checkFlags] repository must be specified")
	}

	return rawInputs{
		Repo:   repo,
		From:   *from,
		To:     *to,
		Years:  *years,
		Months: *months,
		Weeks:  *weeks,
		All:    *all,
	}, nil
}

// simplifies processInput signature
type processedInputs struct {
	Owner string
	Repo  string
	From  time.Time
	To    time.Time
	All   bool
}

// splitting this out makes testing easier
func processInput(p rawInputs) (processedInputs, error) {
	if (p.From != "" || p.To != "") && (p.Weeks != 0 || p.Months != 0 || p.Years != 0) {
		return processedInputs{}, errors.New("[processInput] invalid combination of date range arguments")
	}

	rs := strings.Split(p.Repo, "/")
	if len(rs) != 2 {
		return processedInputs{}, errors.New("[processInput] invalid argument. repo should be given in the form <owner>/<repo>")
	}
	repoOwner, repoName := rs[0], rs[1]

	from, to := time.Time{}, time.Now()

	if p.From != "" {
		var err error // delaration required here so that `from` on next line refers to var in parent scope
		from, err = time.Parse("2006-01-02", p.From)
		if err != nil {
			return processedInputs{}, errors.New("[processInput] invalid `from` value provided. Format: YYYY-MM-DD")
		}
	}

	if p.To != "" {
		var err error // so that `to` on next line` refers to var in parent scope
		to, err = time.Parse("2006-01-02", p.To)
		if err != nil {
			return processedInputs{}, errors.New("[processInput] invalid `to` value provided. Format: YYYY-MM-DD")
		}
	}

	// already check flag combination by this point so no need to worry about from or to
	if p.Weeks != 0 || p.Months != 0 || p.Years != 0 {
		from = time.Now().AddDate(-p.Years, -p.Months, -(p.Weeks * 7))
	}

	return processedInputs{
		Owner: repoOwner,
		Repo:  repoName,
		From:  from,
		To:    to,
		All:   p.All,
	}, nil
}

func printStats(items []app.Contributor) {
	// use tabwriter because some usrenames are long
	// reference: https://blog.robphoenix.com/go/aligning-text-in-go-with-tabwriter/
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, item := range items {
		fmt.Fprintf(w, item.String())
	}
	w.Flush()
}
