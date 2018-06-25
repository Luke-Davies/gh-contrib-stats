package github_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/luke-davies/gh-contrib-stats/pkg/github"
)

var listContributorStatsTestResp = `[
	{
		"author": {
			"login": "Luke-Davies",
			"id": 99999999
		},
		"total": 10,
		"weeks": [
			{
				"w": 1529193600,
				"a": 55,
				"d": 44,
				"c": 3
			},
			{
				"w": 1529798400,
				"a": 33,
				"d": 22,
				"c": 7
			}
		]
	},
	{
		"author": {
			"login": "Ron-Swanson",
			"id": 88888888
		},
		"total": 50,
		"weeks": [
			{
				"w": 1529193600,
				"a": 555,
				"d": 444,
				"c": 40
			},
			{
				"w": 1529798400,
				"a": 333,
				"d": 222,
				"c": 10
			}
		]
	}
]`

var testContributorStats = []github.ContributorStats{
	{
		Author: github.Author{
			Login: "Luke-Davies",
		},
		Weeks: []github.Week{
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
	},
	{
		Author: github.Author{
			Login: "Ron-Swanson",
		},
		Weeks: []github.Week{
			{
				WeekBeginning: 1529193600,
				Additions:     555,
				Deletions:     444,
				Commits:       40,
			},
			{
				WeekBeginning: 1529798400,
				Additions:     333,
				Deletions:     222,
				Commits:       10,
			},
		},
	},
}

func TestListContributorStats(t *testing.T) {
	ts := []struct {
		Name        string
		RespBody    string
		RespHeader  int
		RepoOwner   string
		RepoName    string
		ExpectRes   *[]github.ContributorStats
		ExpectError error
	}{
		{
			Name:       "Happy Path",
			RespBody:   listContributorStatsTestResp,
			RespHeader: http.StatusOK,
			RepoOwner:  "repo-owner",
			RepoName:   "repo-name",
			ExpectRes:  &testContributorStats,
		},
		{
			Name:        "GitHub 202",
			RespBody:    `{}`,
			RespHeader:  http.StatusAccepted,
			RepoOwner:   "repo-owner",
			RepoName:    "repo-name",
			ExpectError: fmt.Errorf("[ListContributorStats] [GitHub Error] GitHub sent a 202, meaning they don't have those stats ready. Try again in a minute"),
		},
		{
			Name:        "GitHub 404",
			RespBody:    `{}`,
			RespHeader:  http.StatusNotFound,
			RepoOwner:   "repo-owner",
			RepoName:    "repo-name",
			ExpectError: fmt.Errorf("[ListContributorStats] [GitHub Error] Did not get successful response from github. Received 404"),
		},
		{
			Name:        "Bad Data",
			RespBody:    `######`,
			RespHeader:  http.StatusOK,
			RepoOwner:   "repo-owner",
			RepoName:    "repo-name",
			ExpectError: fmt.Errorf("[ListContributorStats] Error unmarshalling result from GitHub"),
		},
	}

	for _, tc := range ts {
		t.Run(tc.Name, func(t *testing.T) {
			mockHandler := func(w http.ResponseWriter, r *http.Request) {
				wantURL := fmt.Sprintf("/repos/%s/%s/stats/contributors", tc.RepoOwner, tc.RepoName)
				if r.RequestURI != wantURL {
					t.Errorf("ListContributorStats:\n\nhave request url:\n%+v\n\nwant request url:\n%+v", r.RequestURI, wantURL)
				}
				w.WriteHeader(tc.RespHeader)
				fmt.Fprint(w, tc.RespBody)
			}
			mockServer := httptest.NewServer(http.HandlerFunc(mockHandler))
			defer mockServer.Close()

			client := github.Client{BaseURL: mockServer.URL}
			res, err := client.ListContributorStats(context.Background(), tc.RepoOwner, tc.RepoName)
			if err != nil && tc.ExpectError == nil {
				t.Errorf("ListContributorStats: Unexpected Error: %v", err)
			}

			if tc.ExpectError != nil {
				if err == nil {
					t.Fatal("ListContributorStats: Expected error but received nil")
				}
				if !strings.HasPrefix(err.Error(), tc.ExpectError.Error()) {
					t.Errorf("ListContributorStats:\n\nhave error:\n%+v\n\nwant error that starts with:\n%+v", err.Error(), tc.ExpectError.Error())
				}
			}

			if tc.ExpectRes != nil {
				if !reflect.DeepEqual(res, tc.ExpectRes) {
					t.Errorf("ListContributorStats:\n\nhave result:\n%+v\n\nwant result:\n%+v", res, tc.ExpectRes)
				}
			}
		})
	}
}
