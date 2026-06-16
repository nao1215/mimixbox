// mimixbox/internal/applets/applet_meta_test.go
//
// Tests for the top-level UX surfaces added in issues #781-#784: applet
// subsystem/stability metadata, JSON listing, prefix/subsystem filtering, and
// nearest-name suggestions.
package applets

import (
	"bytes"
	"encoding/json"
	"sort"
	"testing"
)

// TestAppletMetadata verifies every applet carries a non-empty subsystem and a
// known stability value, and spot-checks a few representative defaults.
func TestAppletMetadata(t *testing.T) {
	t.Parallel()

	valid := map[Stability]bool{
		StabilityStable:       true,
		StabilityPartial:      true,
		StabilityGated:        true,
		StabilityExperimental: true,
	}
	for name, a := range Applets {
		if a.Subsystem == "" {
			t.Errorf("applet %q has an empty subsystem", name)
		}
		if !valid[a.Stability] {
			t.Errorf("applet %q has unknown stability %q", name, a.Stability)
		}
	}

	tests := []struct {
		name    string
		sub     string
		stab    Stability
		present bool
	}{
		{name: "cat", sub: "textutils", stab: StabilityStable, present: true},
		{name: "ls", sub: "fileutils", stab: StabilityStable, present: true},
		{name: "passwd", sub: "loginutils", stab: StabilityGated, present: true},
	}
	for _, tt := range tests {
		a, ok := Applets[tt.name]
		if ok != tt.present {
			t.Fatalf("applet %q presence = %v, want %v", tt.name, ok, tt.present)
		}
		if !ok {
			continue
		}
		if a.Subsystem != tt.sub {
			t.Errorf("applet %q subsystem = %q, want %q", tt.name, a.Subsystem, tt.sub)
		}
		if a.Stability != tt.stab {
			t.Errorf("applet %q stability = %q, want %q", tt.name, a.Stability, tt.stab)
		}
	}
}

// TestSuggestApplets checks the nearest-name suggestion algorithm.
func TestSuggestApplets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target string
		want   string // must be the first suggestion
	}{
		{name: "transposed ls", target: "lss", want: "ls"},
		{name: "cat typo", target: "caat", want: "cat"},
		{name: "grep typo", target: "grpe", want: "grep"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SuggestApplets(tt.target, 5)
			if len(got) == 0 {
				t.Fatalf("SuggestApplets(%q) returned no suggestions, want %q first", tt.target, tt.want)
			}
			if got[0] != tt.want {
				t.Errorf("SuggestApplets(%q)[0] = %q, want %q (all: %v)", tt.target, got[0], tt.want, got)
			}
			if len(got) > 5 {
				t.Errorf("SuggestApplets(%q) returned %d suggestions, want <= 5", tt.target, len(got))
			}
		})
	}

	// Wildly different input should produce nothing (above the distance cap).
	if got := SuggestApplets("zzzzzzzzzzzzzzz", 5); len(got) != 0 {
		t.Errorf("SuggestApplets(garbage) = %v, want empty", got)
	}
}

// TestSuggestAppletsSorted asserts suggestions are ordered by distance then name.
func TestSuggestAppletsSorted(t *testing.T) {
	t.Parallel()

	got := SuggestApplets("lss", 5)
	if len(got) == 0 {
		t.Fatal("expected at least one suggestion for 'lss'")
	}
	dists := make([]int, len(got))
	for i, name := range got {
		dists[i] = levenshtein("lss", name)
	}
	if !sort.IntsAreSorted(dists) {
		t.Errorf("suggestions not ordered by distance: %v -> %v", got, dists)
	}
}

// TestListAppletsJSONTo validates the JSON schema and ordering.
func TestListAppletsJSONTo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	if err := ListAppletsJSONTo(&out, ListFilter{}); err != nil {
		t.Fatalf("ListAppletsJSONTo: %v", err)
	}

	var got []jsonApplet
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(got) != len(Applets) {
		t.Fatalf("JSON has %d entries, want %d", len(got), len(Applets))
	}

	// Sorted by name, schema fields populated.
	names := make([]string, len(got))
	for i, e := range got {
		names[i] = e.Name
		if e.Name == "" || e.Subsystem == "" || e.Stability == "" {
			t.Errorf("entry %d has empty required field: %+v", i, e)
		}
		want := Applets[e.Name]
		if e.Synopsis != want.Desc {
			t.Errorf("entry %q synopsis = %q, want %q", e.Name, e.Synopsis, want.Desc)
		}
		if e.Subsystem != want.Subsystem {
			t.Errorf("entry %q subsystem = %q, want %q", e.Name, e.Subsystem, want.Subsystem)
		}
	}
	if !sort.StringsAreSorted(names) {
		t.Error("JSON entries are not sorted by name")
	}

	// cat and ls must be present (used by the ShellSpec contract too).
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	for _, must := range []string{"cat", "ls"} {
		if !found[must] {
			t.Errorf("JSON listing is missing %q", must)
		}
	}
}

// TestFilteredApplets checks prefix and subsystem filtering semantics.
func TestFilteredApplets(t *testing.T) {
	t.Parallel()

	// No filter returns everything.
	if got := len(FilteredApplets(ListFilter{})); got != len(Applets) {
		t.Errorf("empty filter returned %d, want %d", got, len(Applets))
	}

	// Prefix filter: inclusion and exclusion.
	cat := FilteredApplets(ListFilter{Prefix: "cat"})
	if !contains(cat, "cat") {
		t.Errorf("prefix=cat should include cat, got %v", cat)
	}
	if contains(cat, "ls") {
		t.Errorf("prefix=cat should exclude ls, got %v", cat)
	}
	for _, n := range cat {
		if len(n) < 3 || n[:3] != "cat" {
			t.Errorf("prefix=cat returned non-matching name %q", n)
		}
	}

	// Subsystem filter: every result is in the subsystem, and a known member is present.
	tu := FilteredApplets(ListFilter{Subsystem: "textutils"})
	if !contains(tu, "cat") {
		t.Errorf("subsystem=textutils should include cat, got %v", tu)
	}
	if contains(tu, "ls") { // ls is fileutils
		t.Errorf("subsystem=textutils should exclude ls")
	}
	for _, n := range tu {
		if Applets[n].Subsystem != "textutils" {
			t.Errorf("subsystem=textutils returned %q from %q", n, Applets[n].Subsystem)
		}
	}

	// Combined prefix + subsystem.
	combo := FilteredApplets(ListFilter{Prefix: "sha", Subsystem: "textutils"})
	for _, n := range combo {
		if n[:3] != "sha" || Applets[n].Subsystem != "textutils" {
			t.Errorf("combined filter returned non-matching %q", n)
		}
	}

	// A prefix that matches nothing yields an empty slice.
	if got := FilteredApplets(ListFilter{Prefix: "definitely-not-a-prefix"}); len(got) != 0 {
		t.Errorf("non-matching prefix returned %v, want empty", got)
	}
}

func contains(s []string, target string) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}
