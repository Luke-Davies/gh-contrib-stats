package app_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/luke-davies/gh-contrib-stats/pkg/app"
	"github.com/luke-davies/gh-contrib-stats/pkg/github"
)

var testContributorStats = github.ContributorStats{
	Author: github.Author{
		Login: "Luke-Davies",
	},
	Weeks: []github.Week{
		{
			WeekBeginning: 1527984000,
			Additions:     10,
			Deletions:     12,
			Commits:       3,
		},
		{
			WeekBeginning: 1528588800,
			Additions:     11,
			Deletions:     14,
			Commits:       2,
		},
		{
			WeekBeginning: 1529193600,
			Additions:     55,
			Deletions:     44,
			Commits:       3,
		},
		{
			WeekBeginning: 1529798400,
			Additions:     33,
			Deletions:     22,
			Commits:       7,
		},
	},
}

func TestCalcContributions(t *testing.T) {
	// TODO: should really add some boundary tests
	// TODO: test future dates

	ts := []struct {
		Name         string
		GHContrStats github.ContributorStats
		Options      app.CalcContrbutionsOpts
		ExpectRes    app.Contributor
	}{
		{
			Name:         "Full Range",
			GHContrStats: testContributorStats,
			Options:      app.CalcContrbutionsOpts{From: time.Time{}, To: time.Now()},
			ExpectRes: app.Contributor{
				Name: "Luke-Davies",
				Stats: app.Stats{
					Additions: 109,
					Deletions: 92,
					Commits:   15,
				},
			},
		},
		{
			Name:         "No Range",
			GHContrStats: testContributorStats,
			Options:      app.CalcContrbutionsOpts{},
			ExpectRes: app.Contributor{
				Name: "Luke-Davies",
				Stats: app.Stats{
					Additions: 0,
					Deletions: 0,
					Commits:   0,
				},
			},
		},
		{
			Name:         "Single Week",
			GHContrStats: testContributorStats,
			Options: app.CalcContrbutionsOpts{
				// TODO: Should probably rethink how to do this
				From: func() time.Time {
					d, err := time.Parse("2006-01-02", "2018-06-16")
					if err != nil {
						t.Errorf("TestCalcContributions: Problem setting up a test that parses a date")
					}
					return d
				}(),
				To: func() time.Time {
					d, err := time.Parse("2006-01-02", "2018-06-23")
					if err != nil {
						t.Errorf("TestCalcContributions: Problem setting up a test that parses a date")
					}
					return d
				}(),
			},
			ExpectRes: app.Contributor{
				Name: "Luke-Davies",
				Stats: app.Stats{
					Additions: 55,
					Deletions: 44,
					Commits:   3,
				},
			},
		},
	}

	for _, tc := range ts {
		t.Run(tc.Name, func(t *testing.T) {
			res := app.CalcContributions(tc.GHContrStats, tc.Options)
			if !reflect.DeepEqual(res, tc.ExpectRes) {
				t.Errorf("github.ListContributorStats:\n\nhave result:\n%+v\n\nwant result:\n%+v", res, tc.ExpectRes)
			}
		})
	}

}

func TestNormaliseCalcContributionsOpts(t *testing.T) {
	testDate, err := time.Parse("2006-01-02", "2018-06-16")
	if err != nil {
		t.Errorf("TestNormaliseCalcContributionsOpts: Problem parsing a date")
	}
	ts := []struct {
		Name      string
		Input     app.CalcContrbutionsOpts
		ExpectRes app.CalcContrbutionsOpts
	}{
		{
			Name: "No Change",
			Input: app.CalcContrbutionsOpts{
				From: testDate,
				To:   testDate,
			},
			ExpectRes: app.CalcContrbutionsOpts{
				From: testDate,
				To:   testDate,
			},
		},
		{
			Name:  "Should Normalise From",
			Input: app.CalcContrbutionsOpts{},
			ExpectRes: app.CalcContrbutionsOpts{
				To: time.Now(),
			},
		},
	}

	for _, tc := range ts {
		t.Run(tc.Name, func(t *testing.T) {
			res := app.NormaliseCalcContributionsOpts(tc.Input)
			if res.From != tc.ExpectRes.From {
				t.Fatalf("NormaliseCalcContributionsOpts: `From` doesn't match expected.\n\nHave:\n%s\n\nWant:\n%s", res.From, tc.ExpectRes.From)
			}

			if res.To.Format("2006-01-02") != tc.ExpectRes.To.Format("2006-01-02") {
				t.Fatalf(
					"NormaliseCalcContributionsOpts: `To` doesn't match expected.\n\nHave date():\n%s\n\nWant date():\n%s",
					res.To.Format("2006-01-02"), tc.ExpectRes.To.Format("2006-01-02"),
				)
			}
		})
	}
}

func TestValidateCalcContributionsOpts(t *testing.T) {
	earlierDate, err := time.Parse("2006-01-02", "2018-06-16")
	laterDate, err := time.Parse("2006-01-02", "2018-06-17")
	if err != nil {
		t.Errorf("TestNormaliseCalcContributionsOpts: Problem parsing a date")
	}
	ts := []struct {
		Name        string
		Input       app.CalcContrbutionsOpts
		ShouldError bool
	}{
		{
			Name: "Valid",
			Input: app.CalcContrbutionsOpts{
				To:   laterDate,
				From: earlierDate,
			},
		},
		{
			Name: "Same Date",
			Input: app.CalcContrbutionsOpts{
				To:   earlierDate,
				From: earlierDate,
			},
			ShouldError: true,
		},
		{
			Name: "Bad Range",
			Input: app.CalcContrbutionsOpts{
				To:   earlierDate,
				From: laterDate,
			},
			ShouldError: true,
		},
		{
			Name: "`To` in future",
			Input: app.CalcContrbutionsOpts{
				To:   time.Now().AddDate(1, 0, 0),
				From: earlierDate,
			},
			ShouldError: true,
		},
		{
			Name: "`From` in future",
			Input: app.CalcContrbutionsOpts{
				To:   laterDate,
				From: time.Now().AddDate(1, 0, 0),
			},
			ShouldError: true,
		},
	}

	for _, tc := range ts {
		t.Run(tc.Name, func(t *testing.T) {
			err := app.ValidateCalcContributionsOpts(tc.Input)
			if err == nil && tc.ShouldError {
				t.Fatal("Should error but didn't")
			}
			if err != nil && !tc.ShouldError {
				t.Fatal("Should not error but did")
			}
		})
	}
}

func TestFilterContributors(t *testing.T) {
	input := []app.Contributor{
		{
			Name: "Luke-Davies",
			Stats: app.Stats{
				Additions: 109,
				Deletions: 92,
				Commits:   15,
			},
		},
		{
			Name: "Ron-Swanson",
			Stats: app.Stats{
				Additions: 50,
				Deletions: 30,
				Commits:   20,
			},
		},
		{
			Name: "Leslie-Knope",
			Stats: app.Stats{
				Additions: 10,
				Deletions: 10,
				Commits:   10,
			},
		},
		{
			Name: "Andy-Dwyer",
			Stats: app.Stats{
				Additions: 0,
				Deletions: 0,
				Commits:   0,
			},
		},
	}

	res := app.FilterContributors(input, func(c app.Contributor) bool { return c.Stats.Commits > 0 })

	want := input[:3]

	if !reflect.DeepEqual(res, want) {
		t.Errorf("github.FilterContributors:\n\nhave result:\n%+v\n\nwant result:\n%+v", res, want)
	}
}
