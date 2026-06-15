//
// mimixbox/internal/applets/shellutils/leadtime/stat.go
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

import "sort"

// statistics holds the aggregate lead-time metrics, all in minutes.
type statistics struct {
	TotalPR int     `json:"total_pr"`
	Max     float64 `json:"max"`
	Min     float64 `json:"min"`
	Sum     float64 `json:"sum"`
	Average float64 `json:"average"`
	Median  float64 `json:"median"`
}

// applyFilters removes Pull Requests that match the exclusion options. It does
// not filter by merged state; mergedOnly handles that so unmerged PRs can still
// appear in --all listings when the original tool expects them. The returned
// slice preserves input order.
func applyFilters(prs []pullRequest, opts options) []pullRequest {
	excludePR := make(map[int]struct{}, len(opts.excludePR))
	for _, n := range opts.excludePR {
		excludePR[n] = struct{}{}
	}
	excludeUser := make(map[string]struct{}, len(opts.excludeUser))
	for _, u := range opts.excludeUser {
		excludeUser[u] = struct{}{}
	}

	out := make([]pullRequest, 0, len(prs))
	for _, pr := range prs {
		if opts.excludeBot && pr.isBot() {
			continue
		}
		if _, ok := excludePR[pr.Number]; ok {
			continue
		}
		if _, ok := excludeUser[pr.User.Login]; ok {
			continue
		}
		out = append(out, pr)
	}
	return out
}

// mergedOnly returns the Pull Requests that were actually merged (and therefore
// have a well-defined lead time).
func mergedOnly(prs []pullRequest) []pullRequest {
	out := make([]pullRequest, 0, len(prs))
	for _, pr := range prs {
		if pr.isMerged() {
			out = append(out, pr)
		}
	}
	return out
}

// calcStatistics computes the lead-time statistics over the given merged Pull
// Requests. The caller must ensure prs is non-empty and every element is
// merged.
func calcStatistics(prs []pullRequest) statistics {
	values := make([]float64, 0, len(prs))
	for _, pr := range prs {
		values = append(values, pr.leadTimeMinutes())
	}
	return statistics{
		TotalPR: len(values),
		Max:     maxOf(values),
		Min:     minOf(values),
		Sum:     sumOf(values),
		Average: averageOf(values),
		Median:  medianOf(values),
	}
}

// maxOf returns the largest value. values must be non-empty.
func maxOf(values []float64) float64 {
	m := values[0]
	for _, v := range values[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// minOf returns the smallest value. values must be non-empty.
func minOf(values []float64) float64 {
	m := values[0]
	for _, v := range values[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

// sumOf returns the total of values.
func sumOf(values []float64) float64 {
	var s float64
	for _, v := range values {
		s += v
	}
	return s
}

// averageOf returns the arithmetic mean of values. values must be non-empty.
func averageOf(values []float64) float64 {
	return sumOf(values) / float64(len(values))
}

// medianOf returns the median of values. For an even count it averages the two
// middle elements. values must be non-empty. It does not mutate the input.
func medianOf(values []float64) float64 {
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	mid := n / 2
	if n%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}
