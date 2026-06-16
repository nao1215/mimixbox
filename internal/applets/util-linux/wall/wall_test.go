package wall

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func record(line string) []byte {
	b := make([]byte, recordSize)
	binary.LittleEndian.PutUint16(b[typeOffset:], userProcess)
	copy(b[lineOffset:lineOffset+fieldLen], line)
	return b
}

// setup writes a fixture utmp with the given lines and stubs the injectable
// hooks; it returns the captured writes.
func setup(t *testing.T, lines ...string) *map[string]string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "utmp")
	var data []byte
	for _, l := range lines {
		data = append(data, record(l)...)
	}
	if err := os.WriteFile(f, data, 0o644); err != nil {
		t.Fatal(err)
	}

	captured := map[string]string{}
	origU, origN, origS, origW := utmpPath, now, sender, writeTTY
	utmpPath = f
	now = func() time.Time { return time.Date(2026, 6, 9, 21, 0, 0, 0, time.UTC) }
	sender = func() (string, string) { return "alice", "host1" }
	writeTTY = func(line, text string) error { captured[line] = text; return nil }
	t.Cleanup(func() { utmpPath, now, sender, writeTTY = origU, origN, origS, origW })
	return &captured
}

func runWall(t *testing.T, in string, args ...string) {
	t.Helper()
	io := command.IO{In: strings.NewReader(in), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, args); err != nil {
		t.Fatalf("Run error = %v", err)
	}
}

func TestBroadcastToAllTerminals(t *testing.T) {
	captured := setup(t, "pts/0", "pts/1")
	runWall(t, "", "system going down")

	if len(*captured) != 2 {
		t.Fatalf("expected 2 terminals written, got %d", len(*captured))
	}
	for _, line := range []string{"pts/0", "pts/1"} {
		text := (*captured)[line]
		if !strings.Contains(text, "Broadcast message from alice@host1") {
			t.Errorf("%s banner = %q", line, text)
		}
		if !strings.Contains(text, "system going down") {
			t.Errorf("%s missing message: %q", line, text)
		}
		if !strings.Contains(text, "2026") {
			t.Errorf("%s missing timestamp: %q", line, text)
		}
	}
}

func TestMessageFromStdin(t *testing.T) {
	captured := setup(t, "tty1")
	runWall(t, "from stdin\n")
	if !strings.Contains((*captured)["tty1"], "from stdin") {
		t.Errorf("stdin message not broadcast: %q", (*captured)["tty1"])
	}
}

func TestNoTerminals(t *testing.T) {
	captured := setup(t) // no logins
	runWall(t, "", "hi")
	if len(*captured) != 0 {
		t.Errorf("expected no writes, got %v", *captured)
	}
}

func TestHelpExitStatus(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help Run error = %v", err)
	}
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing exit status section = %q", out.String())
	}
}
