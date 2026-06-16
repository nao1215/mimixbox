package who

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestSynopsis(t *testing.T) {
	t.Parallel()
	if New().Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestCString covers both branches: a NUL-terminated field is trimmed at the
// NUL, and a field with no NUL is returned whole.
func TestCString(t *testing.T) {
	t.Parallel()
	if got := cString([]byte("alice\x00\x00")); got != "alice" {
		t.Errorf("cString(NUL-terminated) = %q, want %q", got, "alice")
	}
	full := []byte("noterminator")
	if got := cString(full); got != "noterminator" {
		t.Errorf("cString(no NUL) = %q, want %q", got, "noterminator")
	}
}

// TestParseUtmpShortRecordIsEOF: a trailing partial record is treated as end of
// data, not an error, and full records before it still parse.
func TestParseUtmpShortRecordIsEOF(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	buf.Write(make([]byte, utmpRecordSize)) // one zero record
	buf.Write([]byte{1, 2, 3})              // trailing short read

	entries, err := parseUtmp(&buf)
	if err != nil {
		t.Fatalf("parseUtmp error = %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("entries = %d, want 1 (short trailing record dropped)", len(entries))
	}
}

// errReader returns an error that is neither EOF nor a short read so parseUtmp's
// error path is exercised.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failure") }

func TestParseUtmpReadError(t *testing.T) {
	t.Parallel()
	if _, err := parseUtmp(errReader{}); err == nil {
		t.Error("parseUtmp should propagate a non-EOF read error")
	}
}

// TestFormatUserIdle covers formatUser's -u branch, which appends the host.
func TestFormatUserIdle(t *testing.T) {
	t.Parallel()
	e := Entry{Type: userProcess, User: "bob", Line: "pts/0", Host: "1.2.3.4"}
	got := formatUser(e, options{idle: true})
	if !strings.Contains(got, "bob") || !strings.Contains(got, "pts/0") {
		t.Errorf("formatUser = %q, want user and line", got)
	}
	if !strings.Contains(got, "1.2.3.4") {
		t.Errorf("formatUser(-u) = %q, want host appended", got)
	}
}
