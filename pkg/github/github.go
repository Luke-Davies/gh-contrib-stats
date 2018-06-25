// Package github provides a client to the GitHub API v3.
// Only the method the app needs is implemented (ListContributorStats)
// and only the fields the app is interested in a specified on the structs.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const (
	userAgent    = "gh-contrib-stats"
	acceptHeader = "application/vnd.github.v3+json"
)

// Client represents a client to the Github API v3
type Client struct {
	BaseURL string
}

// ContributorStats represents the contributor stats returned by GitHub
// - BUT only the parts we're interested in.
type ContributorStats struct {
	Author Author `json:"author"`
	Weeks  []Week `json:"weeks"`
}

// Author represents the author returned by GitHub
// - BUT only the parts we're interested in.
type Author struct {
	Login string `json:"login"`
}

// Week represents the weekly stats returned by GitHub
// We rename the fields here because GitHub names are vague
type Week struct {
	WeekBeginning int64 `json:"w"` // sorry 32-bit users
	Additions     int   `json:"a"`
	Deletions     int   `json:"d"`
	Commits       int   `json:"c"`
}

// ListContributorStats will call the GitHub API for the given repo (owner/name) and return the
// Contributor stats given by GitHub.
// GitHub groups commits, additions and deletions by "week beginning".
func (c Client) ListContributorStats(ctx context.Context, repoOwner, repoName string) (*[]ContributorStats, error) {
	h := http.Client{}
	csURL := fmt.Sprintf("%s/repos/%s/%s/stats/contributors", c.BaseURL, repoOwner, repoName)

	req, err := http.NewRequest(http.MethodGet, csURL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "[ListContributorStats] error creating request for url: %s", csURL)
	}

	// overkill for this but good habit
	req = req.WithContext(ctx)

	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept", acceptHeader)

	resp, err := h.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "[ListContributorStats] error sending request")
	}

	if resp.StatusCode == http.StatusAccepted {
		return nil, errors.New("[ListContributorStats] [GitHub Error] GitHub sent a 202, meaning they don't have those stats ready. Try again in a minute")
	}
	// TODO: what about redirects?
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("[ListContributorStats] [GitHub Error] Did not get successful response from github. Received %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var res []ContributorStats
	err = dec.Decode(&res)
	// don't really need to check err here since the next statement would return it anyway
	// but generally a good habit. (NB: if err nil errors.Wrap returns nil)
	if err != nil {
		return nil, errors.Wrap(err, "[ListContributorStats] Error unmarshalling result from GitHub")
	}

	return &res, nil
}
