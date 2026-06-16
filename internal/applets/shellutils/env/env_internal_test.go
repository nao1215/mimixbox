package env

import (
	"reflect"
	"syscall"
	"testing"
)

// TestSplitArgs covers GNU --split-string expansion: plain whitespace splitting
// plus the supported escape sequences. A single -S string must become multiple
// argv entries.
func TestSplitArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "plain whitespace splits into words", in: "printf %s hi", want: []string{"printf", "%s", "hi"}},
		{name: "collapses runs of whitespace", in: "  a   b\tc ", want: []string{"a", "b", "c"}},
		{name: "tab and newline escapes are literal in a word", in: `a\tb\nc`, want: []string{"a\tb\nc"}},
		{name: "backslash-underscore inserts a space without splitting", in: `one\_two`, want: []string{"one two"}},
		{name: "escaped backslash", in: `a\\b`, want: []string{`a\b`}},
		{name: "empty string yields no args", in: "", want: nil},
		{name: "leading comment yields no args", in: "# this is ignored", want: nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := splitArgs(tt.in)
			if err != nil {
				t.Fatalf("splitArgs(%q) error = %v", tt.in, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitArgs(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}

func TestSplitArgsErrors(t *testing.T) {
	t.Parallel()
	for _, in := range []string{`trailing\`, `bad\q`} {
		if _, err := splitArgs(in); err == nil {
			t.Errorf("splitArgs(%q) = nil error, want error", in)
		}
	}
}

// TestParseSignal accepts valid names (with or without SIG, any case) and
// numbers, and rejects bad names.
func TestParseSignal(t *testing.T) {
	t.Parallel()
	valid := map[string]syscall.Signal{
		"TERM":    syscall.SIGTERM,
		"SIGTERM": syscall.SIGTERM,
		"int":     syscall.SIGINT,
		"SIGHUP":  syscall.SIGHUP,
		"9":       syscall.Signal(9),
	}
	for in, want := range valid {
		got, err := parseSignal(in)
		if err != nil {
			t.Errorf("parseSignal(%q) error = %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("parseSignal(%q) = %d, want %d", in, got, want)
		}
	}

	for _, in := range []string{"", "NOPE", "SIGBOGUS", "-1", "0", "abc"} {
		if _, err := parseSignal(in); err == nil {
			t.Errorf("parseSignal(%q) = nil error, want error", in)
		}
	}
}

// TestParseIgnoreSignals validates list parsing: the all-signals sentinel and
// the empty value expand to the catchable set, a mixed list is parsed, and a
// bad name is rejected.
func TestParseIgnoreSignals(t *testing.T) {
	t.Parallel()

	all, err := parseIgnoreSignals(allSignalsSentinel)
	if err != nil {
		t.Fatalf("parseIgnoreSignals(sentinel) error = %v", err)
	}
	if len(all) != len(catchableSignals) {
		t.Errorf("sentinel produced %d signals, want %d", len(all), len(catchableSignals))
	}

	list, err := parseIgnoreSignals("INT,TERM HUP")
	if err != nil {
		t.Fatalf("parseIgnoreSignals(list) error = %v", err)
	}
	want := []syscall.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP}
	if len(list) != len(want) {
		t.Fatalf("list produced %d signals, want %d", len(list), len(want))
	}
	for i, s := range want {
		if list[i] != s {
			t.Errorf("signal[%d] = %v, want %v", i, list[i], s)
		}
	}

	if _, err := parseIgnoreSignals("INT,NOPE"); err == nil {
		t.Error("parseIgnoreSignals with a bad name = nil error, want error")
	}
}
