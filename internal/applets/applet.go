// mimixbox/internal/applets/applet.go
//
// # Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package applets

//go:generate go run github.com/nao1215/mimixbox/cmd/genapplets

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Stability describes how mature/safe an applet is considered. The generator
// derives a sensible default per subsystem (see cmd/genapplets), and the value
// is surfaced both in README generation and in `mimixbox --list`.
type Stability string

const (
	// StabilityStable marks a well-exercised applet safe for everyday use.
	StabilityStable Stability = "stable"
	// StabilityPartial marks an applet that implements only a subset of the
	// behavior of its system counterpart.
	StabilityPartial Stability = "partial"
	// StabilityGated marks a privileged/destructive applet (mounts, raw block
	// devices, login/security surfaces) that typically needs elevated rights.
	StabilityGated Stability = "gated"
	// StabilityExperimental marks an applet still under active development.
	StabilityExperimental Stability = "experimental"
)

type Applet struct {
	// Cmd is the applet itself. The top-level dispatcher runs it through
	// internal/command.Execute with an injected command.IO, so an applet can be
	// dispatched entirely in memory without mutating os.Args or touching the
	// process streams.
	Cmd  command.Command
	Desc string
	// Subsystem is the applet family (textutils, procps, securityutils, ...),
	// derived from the package path by the generator.
	Subsystem string
	// Stability is the applet's maturity classification; see Stability.
	Stability Stability
}

// Applets is the applet table, populated by the generated init in
// applet_registry_gen.go (run `make generate` to refresh it).
var Applets map[string]Applet

// reg builds an Applet entry for a command. The command's own Synopsis becomes
// the listed description, so the two never drift apart.
func reg(c command.Command, subsystem string, stability Stability) Applet {
	return Applet{Cmd: c, Desc: c.Synopsis(), Subsystem: subsystem, Stability: stability}
}

// register adds c to the applet table under its own Name(), tagged with its
// subsystem and stability. The generated init calls this once per applet
// constructor, so a key can never drift from the command it dispatches to.
func register(c command.Command, subsystem string, stability Stability) {
	Applets[c.Name()] = reg(c, subsystem, stability)
}

func HasApplet(target string) bool {
	_, ok := Applets[target]
	return ok
}

// ListFilter narrows which applets `--list` reports. The zero value matches
// every applet, so an unfiltered `--list` is unchanged.
type ListFilter struct {
	// Prefix keeps only applets whose name begins with this string. A trailing
	// "*" glob is stripped by the caller, so both "cat" and "cat*" arrive here
	// as "cat". Empty matches every name.
	Prefix string
	// Subsystem keeps only applets in this subsystem (textutils, procps, ...).
	// Empty matches every subsystem.
	Subsystem string
}

// matches reports whether a applet satisfies the filter.
func (f ListFilter) matches(name string, a Applet) bool {
	if f.Prefix != "" && !strings.HasPrefix(name, f.Prefix) {
		return false
	}
	if f.Subsystem != "" && a.Subsystem != f.Subsystem {
		return false
	}
	return true
}

// FilteredApplets returns the sorted applet names that satisfy f.
func FilteredApplets(f ListFilter) []string {
	var keys []string
	for name, a := range Applets {
		if f.matches(name, a) {
			keys = append(keys, name)
		}
	}
	sort.Strings(keys)
	return keys
}

// ListAppletsTo writes the "name - description" table to w. Only applets that
// satisfy f are shown; pass the zero ListFilter for the full table.
func ListAppletsTo(w io.Writer, f ListFilter) {
	keys := FilteredApplets(f)
	width := 0
	for _, key := range keys {
		if width < len(key) {
			width = len(key)
		}
	}
	format := "%" + strconv.Itoa(width) + "s - %s\n"
	for _, key := range keys {
		_, _ = fmt.Fprintf(w, format, key, Applets[key].Desc)
	}
}

// jsonApplet is the stable, documented shape of one `--list --json` element.
type jsonApplet struct {
	Name      string `json:"name"`
	Synopsis  string `json:"synopsis"`
	Subsystem string `json:"subsystem"`
	Stability string `json:"stability"`
}

// ListAppletsJSONTo writes the applets matching f to w as a JSON array of
// {name, synopsis, subsystem, stability} objects, sorted by name.
func ListAppletsJSONTo(w io.Writer, f ListFilter) error {
	keys := FilteredApplets(f)
	out := make([]jsonApplet, 0, len(keys))
	for _, key := range keys {
		a := Applets[key]
		out = append(out, jsonApplet{
			Name:      key,
			Synopsis:  a.Desc,
			Subsystem: a.Subsystem,
			Stability: string(a.Stability),
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// SuggestApplets returns up to limit applet names nearest to target, closest
// first. Candidates are ranked primarily by Levenshtein distance; ties are
// broken by the length of the common prefix shared with target (longer first),
// then alphabetically. Biasing toward a shared prefix matches user intent — a
// typo like "lss" should suggest "ls" ahead of the equally-distant "less".
// Candidates whose distance exceeds maxDistance are dropped so wildly different
// input does not produce noise; the result is empty when nothing is close enough.
func SuggestApplets(target string, limit int) []string {
	const maxDistance = 3
	type scored struct {
		name   string
		dist   int
		prefix int
	}
	var cands []scored
	for _, name := range SortApplet() {
		d := levenshtein(target, name)
		if d <= maxDistance {
			cands = append(cands, scored{name: name, dist: d, prefix: commonPrefixLen(target, name)})
		}
	}
	sort.SliceStable(cands, func(i, j int) bool {
		if cands[i].dist != cands[j].dist {
			return cands[i].dist < cands[j].dist
		}
		if cands[i].prefix != cands[j].prefix {
			return cands[i].prefix > cands[j].prefix
		}
		return cands[i].name < cands[j].name
	})
	if limit > 0 && len(cands) > limit {
		cands = cands[:limit]
	}
	out := make([]string, 0, len(cands))
	for _, c := range cands {
		out = append(out, c.name)
	}
	return out
}

// commonPrefixLen returns the number of leading runes a and b share.
func commonPrefixLen(a, b string) int {
	ar, br := []rune(a), []rune(b)
	n := 0
	for n < len(ar) && n < len(br) && ar[n] == br[n] {
		n++
	}
	return n
}

// levenshtein returns the edit distance between a and b.
func levenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	la, lb := len(ar), len(br)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = min3(del, ins, sub)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

// ShowAppletsBySpaceSeparatedTo writes the space-separated, wrapped applet
// names to w.
func ShowAppletsBySpaceSeparatedTo(w io.Writer) {
	const wrap = 60
	var b strings.Builder
	lineLen := 0
	for i, key := range SortApplet() {
		if i > 0 {
			if lineLen+1+len(key) > wrap {
				b.WriteByte('\n')
				lineLen = 0
			} else {
				b.WriteByte(' ')
				lineLen++
			}
		}
		b.WriteString(key)
		lineLen += len(key)
	}
	b.WriteByte('\n')
	_, _ = fmt.Fprint(w, b.String())
}

func SortApplet() []string {
	var keys []string
	for applet := range Applets {
		keys = append(keys, applet)
	}
	sort.Strings(keys)
	return keys
}

func longestAppletLength() int {
	max := 0
	for _, key := range SortApplet() {
		if max < len(key) {
			max = len(key)
		}
	}
	return max
}
