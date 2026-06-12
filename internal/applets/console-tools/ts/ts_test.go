package ts

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, args); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func TestPrefixesTimestamp(t *testing.T) {
	on := now
	now = func() time.Time { return time.Date(2026, 6, 12, 14, 30, 5, 0, time.UTC) }
	defer func() { now = on }()
	out := run(t, "hello\nworld\n")
	want := "2026-06-12 14:30:05 hello\n2026-06-12 14:30:05 world\n"
	if out != want {
		t.Errorf("ts =\n%q\nwant\n%q", out, want)
	}
}

func TestRelative(t *testing.T) {
	var calls int
	on := now
	now = func() time.Time {
		calls++
		// start at t=0, then +2s for the first line's print, etc.
		return time.Unix(int64(calls), 0)
	}
	defer func() { now = on }()
	out := run(t, "a\nb\n", "-r")
	// The first line records start, so its elapsed is small; the second is later.
	if !strings.Contains(out, " a\n") || !strings.Contains(out, " b\n") {
		t.Errorf("relative output = %q", out)
	}
	if strings.Contains(out, "2026") {
		t.Errorf("relative mode should not print absolute dates: %q", out)
	}
}

func TestEmptyInput(t *testing.T) {
	if out := run(t, ""); out != "" {
		t.Errorf("empty input should produce no output, got %q", out)
	}
}
