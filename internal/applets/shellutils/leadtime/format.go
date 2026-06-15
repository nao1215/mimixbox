//
// mimixbox/internal/applets/shellutils/leadtime/format.go
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
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// prDetail is the per-PR view emitted with --all. Lead time is in minutes.
type prDetail struct {
	Number          int     `json:"number"`
	State           string  `json:"state"`
	Title           string  `json:"title"`
	User            string  `json:"user"`
	CreatedAt       string  `json:"created_at"`
	MergedAt        string  `json:"merged_at"`
	LeadTimeMinutes float64 `json:"lead_time_minutes"`
}

// report is the stable JSON schema emitted by --json.
type report struct {
	Owner      string     `json:"owner"`
	Repo       string     `json:"repo"`
	Statistics statistics `json:"statistics"`
	PRs        []prDetail `json:"pull_requests,omitempty"`
}

// toDetails converts merged Pull Requests into the per-PR detail view.
func toDetails(prs []pullRequest) []prDetail {
	out := make([]prDetail, 0, len(prs))
	for _, pr := range prs {
		out = append(out, prDetail{
			Number:          pr.Number,
			State:           pr.State,
			Title:           pr.Title,
			User:            pr.User.Login,
			CreatedAt:       fmtTime(pr.CreatedAt),
			MergedAt:        fmtTime(pr.MergedAt),
			LeadTimeMinutes: pr.leadTimeMinutes(),
		})
	}
	return out
}

// fmtTime renders a timestamp in RFC3339 UTC, or "-" when nil.
func fmtTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.UTC().Format(time.RFC3339)
}

// renderText writes the human-readable default output.
func renderText(w io.Writer, opts options, s statistics, merged []pullRequest) {
	_, _ = fmt.Fprintf(w, "Repository           : %s/%s\n", opts.owner, opts.repo)
	_, _ = fmt.Fprintf(w, "Total PR             : %d\n", s.TotalPR)
	_, _ = fmt.Fprintf(w, "Lead Time(Max)       : %.2f[min]\n", s.Max)
	_, _ = fmt.Fprintf(w, "Lead Time(Min)       : %.2f[min]\n", s.Min)
	_, _ = fmt.Fprintf(w, "Lead Time(Sum)       : %.2f[min]\n", s.Sum)
	_, _ = fmt.Fprintf(w, "Lead Time(Ave)       : %.2f[min]\n", s.Average)
	_, _ = fmt.Fprintf(w, "Lead Time(Median)    : %.2f[min]\n", s.Median)

	if opts.all {
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Pull Requests:")
		for _, d := range toDetails(merged) {
			_, _ = fmt.Fprintf(w, "  #%d [%s] %.2f[min] by %s: %s\n",
				d.Number, d.State, d.LeadTimeMinutes, d.User, d.Title)
		}
	}
}

// renderJSON writes the stable JSON schema. With --all the per-PR details are
// included.
func renderJSON(w io.Writer, opts options, s statistics, merged []pullRequest) error {
	r := report{
		Owner:      opts.owner,
		Repo:       opts.repo,
		Statistics: s,
	}
	if opts.all {
		r.PRs = toDetails(merged)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("can't encode JSON: %w", err)
	}
	return nil
}

// renderMarkdown writes a Markdown table of the statistics, and a per-PR table
// when --all is requested.
func renderMarkdown(w io.Writer, opts options, s statistics, merged []pullRequest) error {
	_, _ = fmt.Fprintf(w, "# Lead Time Statistics for %s/%s\n\n", opts.owner, opts.repo)
	_, _ = fmt.Fprintln(w, "| Metric | Value |")
	_, _ = fmt.Fprintln(w, "| --- | --- |")
	_, _ = fmt.Fprintf(w, "| Total PR | %d |\n", s.TotalPR)
	_, _ = fmt.Fprintf(w, "| Lead Time(Max) | %.2f[min] |\n", s.Max)
	_, _ = fmt.Fprintf(w, "| Lead Time(Min) | %.2f[min] |\n", s.Min)
	_, _ = fmt.Fprintf(w, "| Lead Time(Sum) | %.2f[min] |\n", s.Sum)
	_, _ = fmt.Fprintf(w, "| Lead Time(Ave) | %.2f[min] |\n", s.Average)
	_, _ = fmt.Fprintf(w, "| Lead Time(Median) | %.2f[min] |\n", s.Median)

	if opts.all {
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "## Pull Requests")
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "| PR | State | Lead Time[min] | User | Title |")
		_, _ = fmt.Fprintln(w, "| --- | --- | --- | --- | --- |")
		for _, d := range toDetails(merged) {
			_, _ = fmt.Fprintf(w, "| #%d | %s | %.2f | %s | %s |\n",
				d.Number, d.State, d.LeadTimeMinutes, d.User, d.Title)
		}
	}
	return nil
}
