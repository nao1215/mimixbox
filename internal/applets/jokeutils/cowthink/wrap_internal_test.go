package cowthink

import (
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// TestWrapColumnBoundary covers wrap's column<=0 short-circuit and the
// exact-multiple / remainder splitting cases.
func TestWrapColumnBoundary(t *testing.T) {
	t.Parallel()
	if got := wrap("abcdef", 0); got != "abcdef" {
		t.Errorf("wrap(_,0) = %q, want the unmodified source", got)
	}
	if got := wrap("abcdef", -1); got != "abcdef" {
		t.Errorf("wrap(_,-1) = %q, want the unmodified source", got)
	}
	if got := wrap("abcdef", 3); got != "abc\ndef" {
		t.Errorf("wrap(\"abcdef\",3) = %q, want abc\\ndef", got)
	}
	if got := wrap("abcde", 2); got != "ab\ncd\ne" {
		t.Errorf("wrap(\"abcde\",2) = %q, want ab\\ncd\\ne", got)
	}
}

// errReader fails on the first read, modeling broken stdin.
type errReader struct{ err error }

func (e errReader) Read([]byte) (int, error) { return 0, e.err }

// TestReadMessageScannerError verifies readMessage surfaces a non-EOF scan
// error rather than returning a truncated message.
func TestReadMessageScannerError(t *testing.T) {
	t.Parallel()
	io := command.IO{In: errReader{err: errors.New("read fail")}, Out: &strings.Builder{}, Err: &strings.Builder{}}
	if _, err := readMessage(io); err == nil {
		t.Error("readMessage must surface a scanner read error")
	}
}
