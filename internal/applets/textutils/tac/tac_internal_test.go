package tac

import (
	"strings"
	"testing"
)

// TestWriteReversed exercises the record-splitting/reversal helper directly,
// including the empty-input shortcut and a custom separator.
func TestWriteReversed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		sep  string
		want string
	}{
		{"empty", "", "\n", ""},
		{"trailing newline", "a\nb\nc\n", "\n", "c\nb\na\n"},
		{"no trailing newline", "a\nb\nc", "\n", "cb\na\n"},
		{"comma separator", "1,2,3,", ",", "3,2,1,"},
		{"single record", "only", "\n", "only"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var b strings.Builder
			if err := writeReversed(&b, tt.text, tt.sep); err != nil {
				t.Fatalf("writeReversed err = %v", err)
			}
			if got := b.String(); got != tt.want {
				t.Errorf("writeReversed(%q, %q) = %q, want %q", tt.text, tt.sep, got, tt.want)
			}
		})
	}
}

// errWriter fails on the first write so writeReversed's error path is covered.
type errWriter struct{ err error }

func (e errWriter) Write([]byte) (int, error) { return 0, e.err }

func TestWriteReversedWriteError(t *testing.T) {
	t.Parallel()
	werr := writeReversed(errWriter{err: boomError("boom")}, "a\nb\n", "\n")
	if werr == nil {
		t.Fatal("expected a write error")
	}
}

type boomError string

func (b boomError) Error() string { return string(b) }

// TestFirstNonNil verifies an existing error is preserved and a new silent
// failure is produced otherwise.
func TestFirstNonNil(t *testing.T) {
	t.Parallel()
	existing := boomError("kept")
	if got := firstNonNil(existing); got != error(existing) {
		t.Errorf("firstNonNil(existing) = %v, want %v", got, existing)
	}
	if got := firstNonNil(nil); got == nil {
		t.Error("firstNonNil(nil) = nil, want a failure")
	}
}

func TestSynopsis(t *testing.T) {
	t.Parallel()
	if New().Synopsis() == "" {
		t.Error("Synopsis() = empty")
	}
}
