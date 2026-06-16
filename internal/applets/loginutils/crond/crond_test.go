package crond

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestMatchField(t *testing.T) {
	t.Parallel()
	cases := []struct {
		spec     string
		val      int
		min, max int
		want     bool
	}{
		{"*", 30, 0, 59, true},
		{"30", 30, 0, 59, true},
		{"30", 31, 0, 59, false},
		{"1-5", 3, 0, 59, true},
		{"1-5", 6, 0, 59, false},
		{"*/15", 30, 0, 59, true},
		{"*/15", 31, 0, 59, false},
		{"10-20/5", 15, 0, 59, true},
		{"10-20/5", 16, 0, 59, false},
		{"1,3,5", 5, 0, 59, true},
		{"1,3,5", 4, 0, 59, false},
	}
	for _, c := range cases {
		if got := matchField(c.spec, c.val, c.min, c.max); got != c.want {
			t.Errorf("matchField(%q, %d) = %v, want %v", c.spec, c.val, got, c.want)
		}
	}
}

// 2026-06-11 14:30 is a Thursday (weekday 4).
var ref = time.Date(2026, 6, 11, 14, 30, 0, 0, time.UTC)

func TestEntryMatches(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"30 14 * * *":  true,
		"0 14 * * *":   false, // wrong minute
		"*/15 * * * *": true,  // 30 % 15 == 0
		"30 14 11 6 *": true,  // day 11, month 6
		"30 14 * * 4":  true,  // Thursday
		"30 14 * * 1":  false, // Monday
		"30 14 1 * 4":  true,  // dom and dow both restricted -> either matches (Thu)
		"30 14 11 * 1": true,  // dom matches even though dow (Mon) doesn't
		"30 14 1 * 1":  false, // neither dom (1) nor dow (Mon) matches
	}
	for spec, want := range cases {
		e := parseEntries(spec + " /bin/true")[0]
		if got := e.matches(ref); got != want {
			t.Errorf("%q matches = %v, want %v", spec, got, want)
		}
	}
}

func TestSundayAsSeven(t *testing.T) {
	t.Parallel()
	// 2026-06-14 is a Sunday (weekday 0); "* * * * 7" should match it.
	sun := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	e := parseEntries("* * * * 7 cmd")[0]
	if !e.matches(sun) {
		t.Errorf("dow 7 should match Sunday")
	}
}

func TestParseEntries(t *testing.T) {
	t.Parallel()
	text := "# a comment\n\nMAILTO=root\n*/5 * * * * /usr/bin/backup --now\n0 0 * * 0 weekly\n"
	entries := parseEntries(text)
	if len(entries) != 2 {
		t.Fatalf("parsed %d entries, want 2", len(entries))
	}
	if entries[0].command != "/usr/bin/backup --now" {
		t.Errorf("command = %q", entries[0].command)
	}
}

func TestRunDue(t *testing.T) {
	t.Parallel()
	entries := parseEntries("30 14 * * * yes-job\n0 0 * * * no-job")
	var ran []string
	runDue(entries, ref, func(c string) { ran = append(ran, c) })
	if len(ran) != 1 || ran[0] != "yes-job" {
		t.Errorf("ran = %v, want [yes-job]", ran)
	}
}

func TestForegroundRequired(t *testing.T) {
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Errorf("crond without -f should fail")
	}
}

func TestStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		done <- New().Run(ctx, io, []string{"-f"})
	}()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned %v after cancel", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("crond did not stop after cancellation")
	}
}

// TestHelpNotes asserts the --help output documents a Notes section.
func TestHelpNotes(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	if !strings.Contains(out.String(), "Notes:") {
		t.Errorf("--help missing Notes section: %q", out.String())
	}
}
