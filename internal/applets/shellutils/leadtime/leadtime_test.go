//
// mimixbox/internal/applets/shellutils/leadtime/leadtime_test.go
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
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// requireLoopback skips the test when a loopback listen socket cannot be
// created (e.g. a sandbox without networking), since httptest needs one.
func requireLoopback(t *testing.T) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback listen unavailable: %v", err)
	}
	_ = ln.Close()
}

// mustTime parses an RFC3339 timestamp or fails the test.
func mustTime(t *testing.T, s string) *time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("bad time %q: %v", s, err)
	}
	return &tm
}

// pr is a small helper to build a merged pull request created at create and
// merged at merge (both RFC3339), authored by login of the given type.
func pr(t *testing.T, number int, login, typ, create, merge string) pullRequest {
	t.Helper()
	p := pullRequest{Number: number, State: "closed", Title: "PR " + strconv.Itoa(number)}
	p.User.Login = login
	p.User.Type = typ
	p.CreatedAt = mustTime(t, create)
	if merge != "" {
		p.MergedAt = mustTime(t, merge)
	}
	return p
}

func almostEqual(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestCalcStatistics(t *testing.T) {
	// Lead times: 10, 20, 30, 40 minutes.
	prs := []pullRequest{
		pr(t, 1, "alice", "User", "2024-01-01T00:00:00Z", "2024-01-01T00:10:00Z"),
		pr(t, 2, "bob", "User", "2024-01-01T00:00:00Z", "2024-01-01T00:20:00Z"),
		pr(t, 3, "carol", "User", "2024-01-01T00:00:00Z", "2024-01-01T00:30:00Z"),
		pr(t, 4, "dave", "User", "2024-01-01T00:00:00Z", "2024-01-01T00:40:00Z"),
	}
	s := calcStatistics(prs)
	if s.TotalPR != 4 {
		t.Errorf("TotalPR = %d, want 4", s.TotalPR)
	}
	if !almostEqual(s.Max, 40) {
		t.Errorf("Max = %v, want 40", s.Max)
	}
	if !almostEqual(s.Min, 10) {
		t.Errorf("Min = %v, want 10", s.Min)
	}
	if !almostEqual(s.Sum, 100) {
		t.Errorf("Sum = %v, want 100", s.Sum)
	}
	if !almostEqual(s.Average, 25) {
		t.Errorf("Average = %v, want 25", s.Average)
	}
	// Even count median: (20+30)/2 = 25.
	if !almostEqual(s.Median, 25) {
		t.Errorf("Median = %v, want 25", s.Median)
	}
}

func TestMedianOdd(t *testing.T) {
	prs := []pullRequest{
		pr(t, 1, "a", "User", "2024-01-01T00:00:00Z", "2024-01-01T00:10:00Z"),
		pr(t, 2, "b", "User", "2024-01-01T00:00:00Z", "2024-01-01T00:30:00Z"),
		pr(t, 3, "c", "User", "2024-01-01T00:00:00Z", "2024-01-01T00:50:00Z"),
	}
	s := calcStatistics(prs)
	if !almostEqual(s.Median, 30) {
		t.Errorf("Median = %v, want 30", s.Median)
	}
}

func TestMedianDoesNotMutate(t *testing.T) {
	in := []float64{30, 10, 20}
	_ = medianOf(in)
	if in[0] != 30 || in[1] != 10 || in[2] != 20 {
		t.Errorf("medianOf mutated input: %v", in)
	}
}

func TestApplyFilters(t *testing.T) {
	prs := []pullRequest{
		pr(t, 1, "alice", "User", "2024-01-01T00:00:00Z", "2024-01-01T00:10:00Z"),
		pr(t, 2, "dependabot[bot]", "Bot", "2024-01-01T00:00:00Z", "2024-01-01T00:20:00Z"),
		pr(t, 3, "bob", "User", "2024-01-01T00:00:00Z", "2024-01-01T00:30:00Z"),
		pr(t, 4, "carol", "User", "2024-01-01T00:00:00Z", "2024-01-01T00:40:00Z"),
	}

	tests := []struct {
		name string
		opts options
		want []int
	}{
		{"no filter", options{}, []int{1, 2, 3, 4}},
		{"exclude bot", options{excludeBot: true}, []int{1, 3, 4}},
		{"exclude pr", options{excludePR: []int{1, 3}}, []int{2, 4}},
		{"exclude user", options{excludeUser: []string{"bob"}}, []int{1, 2, 4}},
		{"combined", options{excludeBot: true, excludePR: []int{4}, excludeUser: []string{"bob"}}, []int{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyFilters(prs, tt.opts)
			var nums []int
			for _, p := range got {
				nums = append(nums, p.Number)
			}
			if len(nums) != len(tt.want) {
				t.Fatalf("got %v, want %v", nums, tt.want)
			}
			for i := range nums {
				if nums[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", nums, tt.want)
				}
			}
		})
	}
}

func TestMergedOnly(t *testing.T) {
	prs := []pullRequest{
		pr(t, 1, "a", "User", "2024-01-01T00:00:00Z", "2024-01-01T00:10:00Z"), // merged
		pr(t, 2, "b", "User", "2024-01-01T00:00:00Z", ""),                     // unmerged (open/closed)
	}
	// PR 3 has merged_at but missing created_at -> not a valid lead time.
	p3 := pullRequest{Number: 3, State: "closed"}
	p3.MergedAt = mustTime(t, "2024-01-01T00:30:00Z")
	prs = append(prs, p3)

	got := mergedOnly(prs)
	if len(got) != 1 || got[0].Number != 1 {
		var nums []int
		for _, p := range got {
			nums = append(nums, p.Number)
		}
		t.Errorf("mergedOnly = %v, want [1]", nums)
	}
}

func TestParseExcludePR(t *testing.T) {
	got, err := parseExcludePR(" 1, 3 ,19 ")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := []int{1, 3, 19}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if _, err := parseExcludePR("1,x"); err == nil {
		t.Error("expected error for non-numeric PR token")
	}
	if g, _ := parseExcludePR(""); g != nil {
		t.Errorf("empty should be nil, got %v", g)
	}
}

func TestIsBot(t *testing.T) {
	tests := []struct {
		login string
		typ   string
		want  bool
	}{
		{"alice", "User", false},
		{"dependabot[bot]", "Bot", true},
		{"renovate", "User", true},
		{"github-actions", "User", true},
		{"somebot[bot]", "User", true},
		{"human", "User", false},
	}
	for _, tt := range tests {
		p := pullRequest{}
		p.User.Login = tt.login
		p.User.Type = tt.typ
		if got := p.isBot(); got != tt.want {
			t.Errorf("isBot(%q,%q) = %v, want %v", tt.login, tt.typ, got, tt.want)
		}
	}
}

// --- API client tests via httptest ---

// pageJSON is the canned first page: two merged PRs, one unmerged, one bot.
const pageJSON = `[
  {"number":1,"state":"closed","title":"feat one","created_at":"2024-01-01T00:00:00Z","merged_at":"2024-01-01T01:00:00Z","user":{"login":"alice","type":"User"}},
  {"number":2,"state":"open","title":"wip two","created_at":"2024-01-02T00:00:00Z","merged_at":null,"user":{"login":"bob","type":"User"}},
  {"number":3,"state":"closed","title":"deps","created_at":"2024-01-03T00:00:00Z","merged_at":"2024-01-03T00:30:00Z","user":{"login":"dependabot[bot]","type":"Bot"}},
  {"number":4,"state":"closed","title":"feat four","created_at":"2024-01-04T00:00:00Z","merged_at":"2024-01-04T03:00:00Z","user":{"login":"carol","type":"User"}}
]`

// newServer starts an httptest server returning body with status for the pulls
// endpoint, sets a token env, and returns the base URL.
func newServer(t *testing.T, status int, body string) string {
	t.Helper()
	requireLoopback(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Page 2+ returns empty so pagination terminates.
		if r.URL.Query().Get("page") != "1" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// runStat runs `leadtime stat` with token set and --base-url pointed at base.
func runStat(t *testing.T, base string, args ...string) (string, string, error) {
	t.Helper()
	t.Setenv("LT_GITHUB_ACCESS_TOKEN", "fake-token")
	t.Setenv("GITHUB_TOKEN", "")
	full := append([]string{"stat", "--owner=acme", "--repo=demo", "--base-url=" + base}, args...)
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: &bytes.Buffer{}, Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, full)
	return out.String(), errBuf.String(), err
}

func TestStatTextDefault(t *testing.T) {
	base := newServer(t, http.StatusOK, pageJSON)
	out, errOut, err := runStat(t, base)
	if err != nil {
		t.Fatalf("err = %v, stderr = %s", err, errOut)
	}
	// Merged PRs: #1 (60min), #3 (30min), #4 (180min). Sum=270, avg=90, median=60.
	for _, want := range []string{
		"Repository           : acme/demo",
		"Total PR             : 3",
		"Lead Time(Max)       : 180.00[min]",
		"Lead Time(Min)       : 30.00[min]",
		"Lead Time(Sum)       : 270.00[min]",
		"Lead Time(Ave)       : 90.00[min]",
		"Lead Time(Median)    : 60.00[min]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull:\n%s", want, out)
		}
	}
}

func TestStatExcludeBot(t *testing.T) {
	base := newServer(t, http.StatusOK, pageJSON)
	out, _, err := runStat(t, base, "--exclude-bot")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// Without the bot PR #3: merged #1 (60), #4 (180). Total 2.
	if !strings.Contains(out, "Total PR             : 2") {
		t.Errorf("expected 2 PRs after exclude-bot, got:\n%s", out)
	}
}

func TestStatExcludePR(t *testing.T) {
	base := newServer(t, http.StatusOK, pageJSON)
	out, _, err := runStat(t, base, "-P", "1")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// Exclude #1: merged #3, #4 remain. Total 2.
	if !strings.Contains(out, "Total PR             : 2") {
		t.Errorf("expected 2 PRs after exclude-pr, got:\n%s", out)
	}
}

func TestStatExcludeUser(t *testing.T) {
	base := newServer(t, http.StatusOK, pageJSON)
	out, _, err := runStat(t, base, "-U", "carol")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// Exclude carol (#4): merged #1, #3 remain. Total 2.
	if !strings.Contains(out, "Total PR             : 2") {
		t.Errorf("expected 2 PRs after exclude-user, got:\n%s", out)
	}
}

func TestStatJSON(t *testing.T) {
	base := newServer(t, http.StatusOK, pageJSON)
	out, _, err := runStat(t, base, "--json", "--all")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	var r report
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if r.Owner != "acme" || r.Repo != "demo" {
		t.Errorf("owner/repo = %s/%s", r.Owner, r.Repo)
	}
	if r.Statistics.TotalPR != 3 {
		t.Errorf("TotalPR = %d, want 3", r.Statistics.TotalPR)
	}
	if !almostEqual(r.Statistics.Average, 90) {
		t.Errorf("Average = %v, want 90", r.Statistics.Average)
	}
	if len(r.PRs) != 3 {
		t.Errorf("with --all expected 3 PR details, got %d", len(r.PRs))
	}
}

func TestStatJSONNoAllOmitsPRs(t *testing.T) {
	base := newServer(t, http.StatusOK, pageJSON)
	out, _, err := runStat(t, base, "--json")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(out, "pull_requests") {
		t.Errorf("without --all, pull_requests must be omitted:\n%s", out)
	}
}

func TestStatMarkdown(t *testing.T) {
	base := newServer(t, http.StatusOK, pageJSON)
	out, _, err := runStat(t, base, "--markdown", "--all")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	for _, want := range []string{
		"# Lead Time Statistics for acme/demo",
		"| Metric | Value |",
		"| Total PR | 3 |",
		"| Lead Time(Ave) | 90.00[min] |",
		"## Pull Requests",
		"| PR | State | Lead Time[min] | User | Title |",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("markdown missing %q\nfull:\n%s", want, out)
		}
	}
}

func TestStatRateLimit(t *testing.T) {
	requireLoopback(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "1700000000")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	t.Cleanup(srv.Close)
	_, errOut, err := runStat(t, srv.URL)
	if err == nil {
		t.Fatal("expected error on rate limit")
	}
	if !strings.Contains(errOut, "rate limit exceeded") {
		t.Errorf("stderr = %q, want rate limit message", errOut)
	}
}

func TestStatNotFound(t *testing.T) {
	base := newServer(t, http.StatusNotFound, `{"message":"Not Found"}`)
	_, errOut, err := runStat(t, base)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	if !strings.Contains(errOut, "not found") {
		t.Errorf("stderr = %q, want not-found message", errOut)
	}
}

func TestStatMalformed(t *testing.T) {
	base := newServer(t, http.StatusOK, `this is not json`)
	_, errOut, err := runStat(t, base)
	if err == nil {
		t.Fatal("expected error on malformed body")
	}
	if !strings.Contains(errOut, "leadtime:") {
		t.Errorf("stderr = %q, want leadtime error prefix", errOut)
	}
}

func TestStatEmptyRepo(t *testing.T) {
	base := newServer(t, http.StatusOK, `[]`)
	_, errOut, err := runStat(t, base)
	if err == nil {
		t.Fatal("expected error on empty repo")
	}
	if !strings.Contains(errOut, "no merged Pull Requests") {
		t.Errorf("stderr = %q, want no-merged message", errOut)
	}
}

func TestStatPagination(t *testing.T) {
	requireLoopback(t)
	// Page 1 returns 100 merged PRs (full page), page 2 returns 1, page 3 empty.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		w.WriteHeader(http.StatusOK)
		switch page {
		case "1":
			_, _ = w.Write([]byte(buildPage(1, 100)))
		case "2":
			_, _ = w.Write([]byte(buildPage(101, 1)))
		default:
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	t.Cleanup(srv.Close)
	out, _, err := runStat(t, srv.URL)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(out, "Total PR             : 101") {
		t.Errorf("expected 101 PRs across pages, got:\n%s", out)
	}
}

// buildPage returns a JSON array of count merged PRs starting at startNumber,
// each with a 10-minute lead time.
func buildPage(startNumber, count int) string {
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < count; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		n := startNumber + i
		b.WriteString(`{"number":`)
		b.WriteString(strconv.Itoa(n))
		b.WriteString(`,"state":"closed","title":"p","created_at":"2024-01-01T00:00:00Z","merged_at":"2024-01-01T00:10:00Z","user":{"login":"u","type":"User"}}`)
	}
	b.WriteByte(']')
	return b.String()
}

func TestStatMissingToken(t *testing.T) {
	t.Setenv("LT_GITHUB_ACCESS_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: &bytes.Buffer{}, Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, []string{"stat", "--owner=acme", "--repo=demo"})
	if err == nil {
		t.Fatal("expected error when no token is set")
	}
	if !strings.Contains(errBuf.String(), "no GitHub token") {
		t.Errorf("stderr = %q, want missing-token message", errBuf.String())
	}
}

func TestTokenFallbackToGitHubToken(t *testing.T) {
	t.Setenv("LT_GITHUB_ACCESS_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "gh-token")
	if got := resolveToken(); got != "gh-token" {
		t.Errorf("resolveToken = %q, want gh-token", got)
	}
	t.Setenv("LT_GITHUB_ACCESS_TOKEN", "lt-token")
	if got := resolveToken(); got != "lt-token" {
		t.Errorf("resolveToken = %q, want lt-token (LT_ takes precedence)", got)
	}
}

func TestStatMissingOwnerRepo(t *testing.T) {
	t.Setenv("LT_GITHUB_ACCESS_TOKEN", "x")
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: &bytes.Buffer{}, Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, []string{"stat"})
	if err == nil {
		t.Fatal("expected error when owner/repo missing")
	}
	if !strings.Contains(errBuf.String(), "--owner and --repo are required") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestUnknownSubcommand(t *testing.T) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: &bytes.Buffer{}, Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, []string{"bogus", "--owner=a", "--repo=b"})
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !strings.Contains(errBuf.String(), "unknown subcommand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestJSONMarkdownMutuallyExclusive(t *testing.T) {
	t.Setenv("LT_GITHUB_ACCESS_TOKEN", "x")
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: &bytes.Buffer{}, Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, []string{"stat", "--owner=a", "--repo=b", "--json", "--markdown"})
	if err == nil {
		t.Fatal("expected error for --json --markdown together")
	}
	if !strings.Contains(errBuf.String(), "mutually exclusive") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestHelp(t *testing.T) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: &bytes.Buffer{}, Out: out, Err: errBuf}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Usage: leadtime", "Lead time is the elapsed time", "LT_GITHUB_ACCESS_TOKEN", "rate"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q", want)
		}
	}
}
