package script

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func requireSh(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skipf("sh not on PATH: %v", err)
	}
}

func TestScriptRecords(t *testing.T) {
	requireSh(t)
	// Freeze the clock so timing is deterministic.
	orig := clock
	clock = func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }
	defer func() { clock = orig }()

	dir := t.TempDir()
	ts := filepath.Join(dir, "out.txt")
	tm := filepath.Join(dir, "timing")
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := NewScript().Run(context.Background(), io, []string{"-c", "printf hello", "-T", tm, ts}); err != nil {
		t.Fatalf("script error = %v", err)
	}

	data, _ := os.ReadFile(ts)
	s := string(data)
	if !strings.Contains(s, "Script started on") || !strings.Contains(s, "hello") || !strings.Contains(s, "Script done on") {
		t.Errorf("typescript = %q", s)
	}
	timing, _ := os.ReadFile(tm)
	if !strings.Contains(string(timing), "5") { // "hello" is 5 bytes
		t.Errorf("timing = %q", timing)
	}
}

func TestRoundTrip(t *testing.T) {
	requireSh(t)
	origSleep := sleep
	sleep = func(time.Duration) {} // do not actually wait
	defer func() { sleep = origSleep }()

	dir := t.TempDir()
	ts := filepath.Join(dir, "out.txt")
	tm := filepath.Join(dir, "timing")

	rec := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := NewScript().Run(context.Background(), rec, []string{"-c", "printf 'one\\ntwo\\n'", "-T", tm, ts}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	play := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := NewScriptreplay().Run(context.Background(), play, []string{tm, ts}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "one\ntwo\n" {
		t.Errorf("replay = %q, want %q", out.String(), "one\ntwo\n")
	}
}

func TestScriptPropagatesExit(t *testing.T) {
	requireSh(t)
	dir := t.TempDir()
	ts := filepath.Join(dir, "out.txt")
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	err := NewScript().Run(context.Background(), io, []string{"-c", "exit 3", ts})
	var ee *command.ExitError
	if e, ok := err.(*command.ExitError); ok {
		ee = e
	}
	if ee == nil || ee.Code != 3 {
		t.Errorf("err = %v, want exit 3", err)
	}
}

func TestReplayMissingTiming(t *testing.T) {
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := NewScriptreplay().Run(context.Background(), io, []string{"/no/such/timing", "/no/such/ts"}); err == nil {
		t.Errorf("missing timing should fail")
	}
}
