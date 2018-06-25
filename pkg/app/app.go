// Package app contains all application logic for the app
// including the applications models (structs)
package app

import (
	"fmt"
	"time"

	"github.com/luke-davies/gh-contrib-stats/pkg/github"
)

// Contributor is our apps model of a contributor
type Contributor struct {
	Name  string
	Stats Stats
}

// Stats is our apps model of contributor stats
type Stats struct {
	Additions int
	Deletions int
	Commits   int
}

func (s Stats) String() string {
	return fmt.Sprintf(
		"Commits: %d\t Additions: %d\t Deletions: %d\t",
		s.Commits, s.Additions, s.Deletions,
	)
}

func (c Contributor) String() string {
	return fmt.Sprintf("Contributor: %s\t %s\n", c.Name, c.Stats)
}

// CalcContrbutionsOpts contains available options for CalcContributions
// - From: the start of the date range to calculate over
// - To: the end of the date range to calculate over
type CalcContrbutionsOpts struct {
	From time.Time
	To   time.Time
}

// CalcContributions calculates the total stats (commits, additions and deletions) of the
// contributor for the given options.
//
// Options include `from` and `to` which form a date range. Passing these options results in
// data only being calculated for the given range.
// The available GitHub data is grouped by week-beginning where the first day of the week is sunday,
// therefore the range can only be applied against week beginning.
// i.e stats with week-beginnings between the date range are included.
// This means that a date range of a monday to saturday (6 days) will result in no data.
func CalcContributions(contributor github.ContributorStats, options CalcContrbutionsOpts) Contributor {
	res := Contributor{Name: contributor.Author.Login} // initialised with zero values for stats

	for _, w := range contributor.Weeks {
		wb := time.Unix(w.WeekBeginning, 0)
		// No guarantee that weeks are in order so can't stop early :(
		if (options.From.Equal(wb) || wb.After(options.From)) && wb.Before(options.To) {
			res.Stats.Additions += w.Additions
			res.Stats.Deletions += w.Deletions
			res.Stats.Commits += w.Commits
		}
	}
	return res
}

// NormaliseCalcContributionsOpts normalises the given options
func NormaliseCalcContributionsOpts(options CalcContrbutionsOpts) CalcContrbutionsOpts {
	from := options.From // The zero value of time.Time is ok for From.
	var to time.Time     // but zero value for To means we should use Now()
	if options.To.IsZero() {
		to = time.Now()
	} else {
		to = options.To
	}
	return CalcContrbutionsOpts{From: from, To: to}
}

// ValidateCalcContributionsOpts vaidates the given options.
// Returns an error if invalid. Silence is golden.
func ValidateCalcContributionsOpts(options CalcContrbutionsOpts) error {
	if options.From.Equal(options.To) || options.From.After(options.To) || options.From.After(time.Now()) || options.To.After(time.Now()) {
		return fmt.Errorf(
			"invalid date range given. " +
				"`from` date must be before `to` date and neither `from` or `to` can be in the future",
		)
	}
	return nil
}

// FilterContributors returns a new slice of Contributors filtered by the given function
func FilterContributors(cs []Contributor, f func(Contributor) bool) []Contributor {
	res := make([]Contributor, 0)
	for _, c := range cs {
		if f(c) {
			res = append(res, c)
		}
	}
	return res
}
