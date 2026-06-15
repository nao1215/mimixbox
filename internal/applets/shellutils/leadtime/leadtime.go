//
// mimixbox/internal/applets/shellutils/leadtime/leadtime.go
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

// Package leadtime implements the leadtime applet: it calculates GitHub Pull
// Request lead-time statistics for a repository. Lead time here is the elapsed
// time, in minutes, from when a Pull Request is created until it is merged.
// Only merged Pull Requests contribute to the statistics. The applet is a port
// of the archived github.com/nao1215/leadtime project and talks to the
// read-only GitHub REST API.
package leadtime

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the leadtime applet.
type Command struct{}

// New returns a leadtime command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "leadtime" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Calculate GitHub PR lead time statistics" }

// options collects the parsed command-line options for `leadtime stat`.
type options struct {
	owner       string
	repo        string
	all         bool
	json        bool
	markdown    bool
	excludeBot  bool
	excludePR   []int
	excludeUser []string
	baseURL     string
	token       string
}

// Run executes leadtime.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "stat --owner=OWNER --repo=REPO [OPTION]...", stdio.Err).WithHelp(command.Help{
		Description: "Calculate GitHub Pull Request lead-time statistics for a repository.\n\n" +
			"Lead time is the elapsed time, in minutes, from when a Pull Request is\n" +
			"created until it is merged. Only merged Pull Requests are counted; open and\n" +
			"closed-but-unmerged Pull Requests are ignored. For the matching Pull\n" +
			"Requests leadtime reports the total count and the maximum, minimum, sum,\n" +
			"average, and median lead time. Use --all to also print per-PR details.\n\n" +
			"The only subcommand is 'stat'. --owner and --repo are required.\n\n" +
			"Authentication: leadtime reads a GitHub access token from the\n" +
			"LT_GITHUB_ACCESS_TOKEN environment variable, falling back to GITHUB_TOKEN.\n" +
			"A token is required; unauthenticated requests are rejected with a\n" +
			"deterministic error. Only read-only REST API calls are made.\n\n" +
			"Rate limits: the GitHub REST API limits authenticated requests (5000/hour\n" +
			"by default). Large repositories are paginated, so a single run may issue\n" +
			"several requests. A rate-limit response is reported as a deterministic\n" +
			"error rather than partial output.",
		Examples: []command.Example{
			{Command: "leadtime stat --owner=nao1215 --repo=sqly", Explain: "Print text statistics for nao1215/sqly."},
			{Command: "leadtime stat --owner=nao1215 --repo=sqly --all", Explain: "Also print per-PR lead-time details."},
			{Command: "leadtime stat --owner=nao1215 --repo=sqly --json", Explain: "Emit machine-readable JSON."},
			{Command: "leadtime stat --owner=nao1215 --repo=sqly --markdown", Explain: "Emit a Markdown table."},
			{Command: "leadtime stat --owner=acme --repo=demo -B -P 1,3 -U bob", Explain: "Exclude bots, PR #1 and #3, and user bob."},
			{Command: "leadtime stat --owner=acme --repo=demo --base-url=http://127.0.0.1:8080", Explain: "Target a GitHub Enterprise or test server."},
		},
		ExitStatus: "0  statistics were calculated and printed.\n1  bad usage, missing token, API error, or no merged Pull Requests.",
		Notes: []string{
			"Token is taken from LT_GITHUB_ACCESS_TOKEN, then GITHUB_TOKEN.",
			"--base-url targets GitHub Enterprise or a local test server; it defaults to https://api.github.com.",
			"Only read-only REST API calls are issued; leadtime never mutates GitHub state.",
			"--json and --markdown are mutually exclusive.",
		},
	})

	owner := fs.String("owner", "", "Repository owner (user or organization) [required]")
	repo := fs.String("repo", "", "Repository name [required]")
	all := fs.BoolP("all", "a", false, "Show per-PR lead-time details in addition to statistics")
	jsonOut := fs.Bool("json", false, "Emit statistics as JSON")
	markdownOut := fs.Bool("markdown", false, "Emit statistics as a Markdown table")
	excludeBot := fs.BoolP("exclude-bot", "B", false, "Exclude Pull Requests opened by bots")
	excludePR := fs.StringP("exclude-pr", "P", "", "Comma-separated PR numbers to exclude (e.g. 1,3,19)")
	excludeUser := fs.StringP("exclude-user", "U", "", "Comma-separated user logins to exclude (e.g. alice,bob)")
	baseURL := fs.String("base-url", defaultBaseURL, "GitHub REST API base URL (for GitHub Enterprise or test servers)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: missing subcommand: expected 'stat'\n", c.Name())
		_, _ = fmt.Fprintf(stdio.Err, "Try '%s --help' for more information.\n", c.Name())
		return command.SilentFailure()
	}
	if operands[0] != "stat" {
		_, _ = fmt.Fprintf(stdio.Err, "%s: unknown subcommand '%s'; only 'stat' is supported\n", c.Name(), operands[0])
		_, _ = fmt.Fprintf(stdio.Err, "Try '%s --help' for more information.\n", c.Name())
		return command.SilentFailure()
	}
	if len(operands) > 1 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: extra operand '%s'\n", c.Name(), operands[1])
		_, _ = fmt.Fprintf(stdio.Err, "Try '%s --help' for more information.\n", c.Name())
		return command.SilentFailure()
	}

	if *owner == "" || *repo == "" {
		_, _ = fmt.Fprintf(stdio.Err, "%s: --owner and --repo are required\n", c.Name())
		_, _ = fmt.Fprintf(stdio.Err, "Try '%s --help' for more information.\n", c.Name())
		return command.SilentFailure()
	}
	if *jsonOut && *markdownOut {
		_, _ = fmt.Fprintf(stdio.Err, "%s: --json and --markdown are mutually exclusive\n", c.Name())
		return command.SilentFailure()
	}

	prNumbers, err := parseExcludePR(*excludePR)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}

	opts := options{
		owner:       *owner,
		repo:        *repo,
		all:         *all,
		json:        *jsonOut,
		markdown:    *markdownOut,
		excludeBot:  *excludeBot,
		excludePR:   prNumbers,
		excludeUser: parseExcludeUser(*excludeUser),
		baseURL:     *baseURL,
		token:       resolveToken(),
	}

	return c.run(ctx, stdio, opts)
}

// run performs the real work once options have been validated.
func (c *Command) run(ctx context.Context, stdio command.IO, opts options) error {
	if opts.token == "" {
		_, _ = fmt.Fprintf(stdio.Err, "%s: no GitHub token found; set LT_GITHUB_ACCESS_TOKEN or GITHUB_TOKEN\n", c.Name())
		return command.SilentFailure()
	}

	client := newClient(opts.baseURL, opts.token)
	prs, err := client.fetchPullRequests(ctx, opts.owner, opts.repo)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}

	filtered := applyFilters(prs, opts)
	merged := mergedOnly(filtered)
	if len(merged) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: no merged Pull Requests found for %s/%s\n", c.Name(), opts.owner, opts.repo)
		return command.SilentFailure()
	}

	stats := calcStatistics(merged)

	switch {
	case opts.json:
		return renderJSON(stdio.Out, opts, stats, merged)
	case opts.markdown:
		return renderMarkdown(stdio.Out, opts, stats, merged)
	default:
		renderText(stdio.Out, opts, stats, merged)
		return nil
	}
}

// resolveToken returns the GitHub access token, preferring
// LT_GITHUB_ACCESS_TOKEN and falling back to GITHUB_TOKEN.
func resolveToken() string {
	if t := strings.TrimSpace(os.Getenv("LT_GITHUB_ACCESS_TOKEN")); t != "" {
		return t
	}
	return strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
}

// parseExcludePR parses a comma-separated list of PR numbers ("1,3,19") into a
// slice of ints. An empty string yields nil. A non-numeric token is an error.
func parseExcludePR(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	var out []int
	for _, tok := range strings.Split(s, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		n, err := strconv.Atoi(tok)
		if err != nil {
			return nil, fmt.Errorf("invalid --exclude-pr value %q: must be a PR number", tok)
		}
		out = append(out, n)
	}
	return out, nil
}

// parseExcludeUser parses a comma-separated list of user logins ("alice,bob")
// into a slice. An empty string yields nil.
func parseExcludeUser(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	for _, tok := range strings.Split(s, ",") {
		tok = strings.TrimSpace(tok)
		if tok != "" {
			out = append(out, tok)
		}
	}
	return out
}
