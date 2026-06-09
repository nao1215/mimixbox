package last

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func record(typ int16, line, user, host string, tv int32) []byte {
	b := make([]byte, recordSize)
	binary.LittleEndian.PutUint16(b[typeOffset:], uint16(typ))
	copy(b[lineOffset:lineOffset+fieldLen], line)
	copy(b[userOffset:userOffset+fieldLen], user)
	copy(b[hostOffset:hostOffset+hostLen], host)
	binary.LittleEndian.PutUint32(b[tvSecOffset:], uint32(tv))
	return b
}

func writeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "wtmp")
	var data []byte
	data = append(data, record(bootTime, "~", "reboot", "5.15", 1699990000)...)
	data = append(data, record(userProcess, "pts/0", "alice", "10.0.0.1", 1700000000)...)
	data = append(data, record(deadProcess, "pts/0", "", "", 1700003600)...)
	data = append(data, record(userProcess, "pts/1", "bob", "10.0.0.2", 1700007200)...)
	if err := os.WriteFile(f, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}

func run(t *testing.T, args ...string) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, args); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return out.String()
}

func TestPairing(t *testing.T) {
	t.Parallel()
	out := run(t, writeFixture(t))
	if !strings.Contains(out, "bob") || !strings.Contains(out, "still logged in") {
		t.Errorf("bob should still be logged in: %q", out)
	}
	if !strings.Contains(out, "alice") || !strings.Contains(out, "(01:00)") {
		t.Errorf("alice session should be one hour: %q", out)
	}
	if !strings.Contains(out, "reboot") || !strings.Contains(out, "system boot") {
		t.Errorf("reboot entry missing: %q", out)
	}
	if !strings.Contains(out, "wtmp begins") {
		t.Errorf("footer missing: %q", out)
	}
}

func TestNewestFirst(t *testing.T) {
	t.Parallel()
	out := run(t, writeFixture(t))
	lines := strings.Split(out, "\n")
	// bob (newest login) appears before alice (older).
	bobAt, aliceAt := -1, -1
	for i, l := range lines {
		if strings.Contains(l, "bob") {
			bobAt = i
		}
		if strings.Contains(l, "alice") {
			aliceAt = i
		}
	}
	if bobAt < 0 || aliceAt < 0 || bobAt > aliceAt {
		t.Errorf("expected bob before alice, got bob@%d alice@%d", bobAt, aliceAt)
	}
}

func TestCountLimit(t *testing.T) {
	t.Parallel()
	out := run(t, "-n", "1", writeFixture(t))
	// Only the newest entry (bob) is shown; older ones are omitted.
	if !strings.Contains(out, "bob") {
		t.Errorf("-n 1 should show the newest entry (bob): %q", out)
	}
	if strings.Contains(out, "alice") || strings.Contains(out, "reboot") {
		t.Errorf("-n 1 should omit older entries: %q", out)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"/no/such/wtmp"}); err == nil {
		t.Errorf("missing wtmp should fail")
	}
}
