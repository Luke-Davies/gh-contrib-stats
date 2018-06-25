package main

import (
	"fmt"
	"testing"
	"time"
)

func TestProcessInput(t *testing.T) {
	testDateStr := "2018-07-24"
	testDate, err := time.Parse("2006-01-02", testDateStr)
	if err != nil {
		t.Fatal("[TestProcessInput] Something went wrong setting up test dates")
	}

	ts := []struct {
		Name      string
		Input     rawInputs
		ExpectRes processedInputs
		ExpectErr error
	}{
		{
			Name: "From and To",
			Input: rawInputs{
				Repo: "test-owner/test-repo",
				From: testDateStr,
				To:   testDateStr,
			},
			ExpectRes: processedInputs{
				Owner: "test-owner",
				Repo:  "test-repo",
				From:  testDate,
				To:    testDate,
			},
		},
		{
			Name: "just From",
			Input: rawInputs{
				Repo: "test-owner/test-repo",
				From: testDateStr,
			},
			ExpectRes: processedInputs{
				Owner: "test-owner",
				Repo:  "test-repo",
				From:  testDate,
				To:    time.Now(),
			},
		},
		{
			Name: "just To",
			Input: rawInputs{
				Repo: "test-owner/test-repo",
				To:   testDateStr,
			},
			ExpectRes: processedInputs{
				Owner: "test-owner",
				Repo:  "test-repo",
				From:  time.Time{},
				To:    testDate,
			},
		},
		{
			Name: "just Weeks",
			Input: rawInputs{
				Repo:  "test-owner/test-repo",
				Weeks: 2,
			},
			ExpectRes: processedInputs{
				Owner: "test-owner",
				Repo:  "test-repo",
				From:  time.Now().AddDate(0, 0, -14),
				To:    time.Now(),
			},
		},
		{
			Name: "just months",
			Input: rawInputs{
				Repo:   "test-owner/test-repo",
				Months: 1,
			},
			ExpectRes: processedInputs{
				Owner: "test-owner",
				Repo:  "test-repo",
				From:  time.Now().AddDate(0, -1, 0),
				To:    time.Now(),
			},
		},
		{
			Name: "just years",
			Input: rawInputs{
				Repo:  "test-owner/test-repo",
				Years: 3,
			},
			ExpectRes: processedInputs{
				Owner: "test-owner",
				Repo:  "test-repo",
				From:  time.Now().AddDate(-3, 0, 0),
				To:    time.Now(),
			},
		},
		{
			Name: "years, months, weeks",
			Input: rawInputs{
				Repo:   "test-owner/test-repo",
				Years:  3,
				Months: 2,
				Weeks:  1,
			},
			ExpectRes: processedInputs{
				Owner: "test-owner",
				Repo:  "test-repo",
				From:  time.Now().AddDate(-3, -2, -7),
				To:    time.Now(),
			},
		},
		{
			Name: "None",
			Input: rawInputs{
				Repo: "test-owner/test-repo",
			},
			ExpectRes: processedInputs{
				Owner: "test-owner",
				Repo:  "test-repo",
				From:  time.Time{},
				To:    time.Now(),
			},
		},
		{
			Name: "invalid combo From and Weeks",
			Input: rawInputs{
				Repo:  "test-owner/test-repo",
				From:  testDateStr,
				Weeks: 1,
			},
			ExpectErr: fmt.Errorf("[processInput] invalid combination of date range arguments"),
		},
		{
			Name: "invalid combo From and Months",
			Input: rawInputs{
				Repo:   "test-owner/test-repo",
				From:   testDateStr,
				Months: 1,
			},
			ExpectErr: fmt.Errorf("[processInput] invalid combination of date range arguments"),
		},
		{
			Name: "invalid combo From and Years",
			Input: rawInputs{
				Repo:  "test-owner/test-repo",
				From:  testDateStr,
				Years: 1,
			},
			ExpectErr: fmt.Errorf("[processInput] invalid combination of date range arguments"),
		},
		{
			Name: "invalid combo To and Weeks",
			Input: rawInputs{
				Repo:  "test-owner/test-repo",
				To:    testDateStr,
				Weeks: 1,
			},
			ExpectErr: fmt.Errorf("[processInput] invalid combination of date range arguments"),
		},
		{
			Name: "invalid combo To and Months",
			Input: rawInputs{
				Repo:   "test-owner/test-repo",
				To:     testDateStr,
				Months: 1,
			},
			ExpectErr: fmt.Errorf("[processInput] invalid combination of date range arguments"),
		},
		{
			Name: "invalid combo To and Years",
			Input: rawInputs{
				Repo:  "test-owner/test-repo",
				To:    testDateStr,
				Years: 1,
			},
			ExpectErr: fmt.Errorf("[processInput] invalid combination of date range arguments"),
		},
		{
			Name: "invalid repo",
			Input: rawInputs{
				Repo: "test-repo",
			},
			ExpectErr: fmt.Errorf("[processInput] invalid argument. repo should be given in the form <owner>/<repo>"),
		},
		{
			Name: "invalid From",
			Input: rawInputs{
				Repo: "test-owner/test-repo",
				From: "blam",
			},
			ExpectErr: fmt.Errorf("[processInput] invalid `from` value provided. Format: YYYY-MM-DD"),
		},
		{
			Name: "invalid To",
			Input: rawInputs{
				Repo: "test-owner/test-repo",
				To:   "blam",
			},
			ExpectErr: fmt.Errorf("[processInput] invalid `to` value provided. Format: YYYY-MM-DD"),
		},
	}

	for _, tc := range ts {
		t.Run(tc.Name, func(t *testing.T) {
			res, err := processInput(tc.Input)

			if err != nil && tc.ExpectErr == nil {
				t.Fatalf("Unexpected Error: %s", err.Error())
			}

			if tc.ExpectErr != nil {
				if err == nil {
					t.Fatal("Expected error but received nil")
				}
				if err.Error() != tc.ExpectErr.Error() {
					t.Fatalf("Have `err`: %s want:%s", err.Error(), tc.ExpectErr.Error())
				}
			}

			if res.Owner != tc.ExpectRes.Owner {
				t.Fatalf("Have `Owner`: %s want:%s", res.Owner, tc.ExpectRes.Owner)
			}

			if res.Repo != tc.ExpectRes.Repo {
				t.Fatalf("Have `Repo`: %s want:%s", res.Repo, tc.ExpectRes.Repo)
			}

			if res.From.Format("2006-01-02") != tc.ExpectRes.From.Format("2006-01-02") {
				t.Fatalf("Have `From`: %s want:%s", res.From.Format("2006-01-02"), tc.ExpectRes.From.Format("2006-01-02"))
			}

			if res.To.Format("2006-01-02") != tc.ExpectRes.To.Format("2006-01-02") {
				t.Fatalf("Have `To`: %s want:%s", res.To.Format("2006-01-02"), tc.ExpectRes.To.Format("2006-01-02"))
			}

			if res.All != tc.ExpectRes.All {
				t.Fatalf("Have `All`: %t want:%t", res.All, tc.ExpectRes.All)
			}
		})
	}
}

// TODO: parseInput is a pain to test because of errors about parsing flags twice.
// Will omit parseInput tests for now but in future should rewrite it to use a more
// GNU-like command line parser which might not have the same testing issues

// TODO: printStats tests (omitted in the interest of time).
