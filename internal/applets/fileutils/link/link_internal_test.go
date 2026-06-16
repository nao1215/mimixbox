package link

import (
	"errors"
	"os"
	"testing"
)

// TestUnwrapLinkError checks that unwrap reduces an *os.LinkError to its
// underlying Err, dropping the file names already printed by the command.
func TestUnwrapLinkError(t *testing.T) {
	t.Parallel()
	inner := errors.New("file exists")
	le := &os.LinkError{Op: "link", Old: "a", New: "b", Err: inner}
	if got := unwrap(le); got != inner {
		t.Errorf("unwrap(LinkError) = %v, want %v", got, inner)
	}
}

// TestUnwrapPassThrough checks that a non-LinkError is returned unchanged.
func TestUnwrapPassThrough(t *testing.T) {
	t.Parallel()
	plain := errors.New("some other error")
	if got := unwrap(plain); got != plain {
		t.Errorf("unwrap(plain) = %v, want %v", got, plain)
	}
}
