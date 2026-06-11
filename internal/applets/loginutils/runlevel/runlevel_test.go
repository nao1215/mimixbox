package runlevel

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

// utmpWithRunLevel writes a utmp fixture containing a RUN_LVL record encoding
// the given previous and current runlevel bytes (prev 0 means "since boot").
func utmpWithRunLevel(t *testing.T, recType uint16, prev, cur byte) string {
	t.Helper()
	rec := make([]byte, recordSize)
	binary.LittleEndian.PutUint16(rec[typeOffset:], recType)
	binary.LittleEndian.PutUint32(rec[pidOffset:], uint32(prev)<<8|uint32(cur))
	p := filepath.Join(t.TempDir(), "utmp")
	if err := os.WriteFile(p, rec, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func run(t *testing.T, path string) (string, error) {
	t.Helper()
	orig := utmpPath
	utmpPath = path
	defer func() { utmpPath = orig }()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, nil)
	return strings.TrimSpace(out.String()), err
}

func TestSinceBoot(t *testing.T) {
	out, err := run(t, utmpWithRunLevel(t, runLevel, 0, '5'))
	if err != nil {
		t.Fatal(err)
	}
	if out != "N 5" {
		t.Errorf("runlevel = %q, want \"N 5\"", out)
	}
}

func TestPreviousRunlevel(t *testing.T) {
	out, err := run(t, utmpWithRunLevel(t, runLevel, '3', '5'))
	if err != nil {
		t.Fatal(err)
	}
	if out != "3 5" {
		t.Errorf("runlevel = %q, want \"3 5\"", out)
	}
}

func TestNoRunLevelRecord(t *testing.T) {
	// A record of a different type must be ignored.
	out, err := run(t, utmpWithRunLevel(t, 7 /* USER_PROCESS */, 0, '5'))
	if err == nil {
		t.Errorf("missing run-level record should fail")
	}
	if out != "unknown" {
		t.Errorf("output = %q, want unknown", out)
	}
}

func TestMissingUtmp(t *testing.T) {
	out, err := run(t, "/no/such/utmp")
	if err == nil {
		t.Errorf("a missing utmp should fail")
	}
	if out != "unknown" {
		t.Errorf("output = %q, want unknown", out)
	}
}
