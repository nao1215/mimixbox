//
// mimixbox/internal/applets/shellutils/leadtime/client.go
//
// Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package leadtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// defaultBaseURL is the public GitHub REST API root.
const defaultBaseURL = "https://api.github.com"

// perPage is the page size requested from the GitHub pulls endpoint. GitHub
// caps this at 100.
const perPage = 100

// pullRequest is the subset of the GitHub Pull Request REST representation that
// leadtime needs to compute lead-time statistics.
type pullRequest struct {
	Number    int        `json:"number"`
	State     string     `json:"state"`
	Title     string     `json:"title"`
	CreatedAt *time.Time `json:"created_at"`
	MergedAt  *time.Time `json:"merged_at"`
	User      struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"user"`
}

// isMerged reports whether the Pull Request was merged (it has a merged_at
// timestamp and a created_at to measure from).
func (p pullRequest) isMerged() bool {
	return p.MergedAt != nil && p.CreatedAt != nil && !p.MergedAt.IsZero() && !p.CreatedAt.IsZero()
}

// isBot reports whether the Pull Request author is a GitHub bot account. GitHub
// marks bot accounts with a user type of "Bot"; well-known bot logins are also
// recognized as a fallback.
func (p pullRequest) isBot() bool {
	if strings.EqualFold(p.User.Type, "Bot") {
		return true
	}
	login := strings.ToLower(p.User.Login)
	return strings.HasSuffix(login, "[bot]") ||
		login == "dependabot" || login == "renovate" || login == "github-actions"
}

// leadTimeMinutes returns the lead time in minutes, from PR creation to merge.
// It is only meaningful when isMerged reports true.
func (p pullRequest) leadTimeMinutes() float64 {
	return p.MergedAt.Sub(*p.CreatedAt).Minutes()
}

// client talks to the read-only GitHub REST API.
type client struct {
	baseURL string
	token   string
	http    *http.Client
}

// newClient builds a client for baseURL authenticating with token. A trailing
// slash on baseURL is trimmed so endpoint construction is predictable.
func newClient(baseURL, token string) *client {
	return &client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// fetchPullRequests retrieves every Pull Request (state=all) for owner/repo,
// following pagination until an empty page is returned. It returns a
// deterministic error on rate-limit, non-200, or malformed responses.
func (c *client) fetchPullRequests(ctx context.Context, owner, repo string) ([]pullRequest, error) {
	var all []pullRequest
	for page := 1; ; page++ {
		batch, err := c.fetchPage(ctx, owner, repo, page)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		all = append(all, batch...)
		if len(batch) < perPage {
			break
		}
	}
	return all, nil
}

// fetchPage retrieves a single page of Pull Requests.
func (c *client) fetchPage(ctx context.Context, owner, repo string, page int) ([]pullRequest, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/pulls?state=all&per_page=%d&page=%d",
		c.baseURL, owner, repo, perPage, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("can't build request for GitHub: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't get response from GitHub: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read response from GitHub: %w", err)
	}

	if err := checkStatus(resp, body); err != nil {
		return nil, err
	}

	var prs []pullRequest
	if err := json.Unmarshal(body, &prs); err != nil {
		return nil, fmt.Errorf("can't parse GitHub response (is owner/repo correct?): %w", err)
	}
	return prs, nil
}

// checkStatus turns a non-success HTTP response into a deterministic error,
// distinguishing rate-limit responses (so callers can recognize them) from
// other failures.
func checkStatus(resp *http.Response, body []byte) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// GitHub signals primary rate-limit exhaustion with 403/429 and an
	// X-RateLimit-Remaining of 0.
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	if (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests) && remaining == "0" {
		return fmt.Errorf("GitHub API rate limit exceeded%s", resetHint(resp))
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("GitHub API authentication failed (status 401): check your token")
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("repository not found (status 404): check --owner and --repo")
	}

	msg := strings.TrimSpace(string(body))
	if len(msg) > 200 {
		msg = msg[:200]
	}
	return fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, msg)
}

// resetHint formats a human-friendly hint about when the rate limit resets,
// derived from the X-RateLimit-Reset header (a Unix timestamp). It returns an
// empty string when the header is absent or unparsable.
func resetHint(resp *http.Response) string {
	reset := resp.Header.Get("X-RateLimit-Reset")
	if reset == "" {
		return ""
	}
	sec, err := strconv.ParseInt(reset, 10, 64)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("; resets at %s", time.Unix(sec, 0).UTC().Format(time.RFC3339))
}
