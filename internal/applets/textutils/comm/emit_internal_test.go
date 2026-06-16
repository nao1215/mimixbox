package comm

import (
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// failingWriter fails every write, modeling a broken stdout (e.g. a closed
// pipe), so the write-error branches of emit and compare are reachable.
type failingWriter struct{ err error }

func (f failingWriter) Write([]byte) (int, error) { return 0, f.err }

func errIO(w *failingWriter) command.IO {
	return command.IO{In: strings.NewReader(""), Out: w, Err: &strings.Builder{}}
}

// TestEmitWriteError verifies emit converts a write failure into a command
// failure, and that a suppressed column (show=false) writes nothing and never
// errors even on a broken writer.
func TestEmitWriteError(t *testing.T) {
	t.Parallel()
	boom := errors.New("broken pipe")
	io := errIO(&failingWriter{err: boom})

	if err := emit(io, true, "", "line"); err == nil {
		t.Error("emit must surface a write error when the column is shown")
	}
	if err := emit(io, false, "", "line"); err != nil {
		t.Errorf("emit on a suppressed column must not error, got %v", err)
	}
}

// TestCompareWriteError verifies compare stops and reports the first write
// failure rather than silently continuing through the merge.
func TestCompareWriteError(t *testing.T) {
	t.Parallel()
	io := errIO(&failingWriter{err: errors.New("broken pipe")})
	c := New()

	// Unique-to-first, common, and unique-to-second lines each hit a different
	// emit call site; any one failing must abort compare.
	a := []string{"apple", "banana"}
	b := []string{"banana", "date"}
	if err := c.compare(io, a, b, true, true, true); err == nil {
		t.Error("compare must surface a write error")
	}
}

// TestCompareTailWriteError drives the trailing loops (after one slice is
// exhausted) into a write failure. With only column 1 shown and b a strict
// prefix of a, the leftover a entries are emitted in the first tail loop; with
// only column 2 shown and a a strict prefix of b, the second tail loop runs.
func TestCompareTailWriteError(t *testing.T) {
	t.Parallel()
	c := New()

	io := errIO(&failingWriter{err: errors.New("broken pipe")})
	if err := c.compare(io, []string{"a", "b", "c"}, []string{"a"}, true, false, false); err == nil {
		t.Error("compare must surface a write error from the first tail loop")
	}

	io = errIO(&failingWriter{err: errors.New("broken pipe")})
	if err := c.compare(io, []string{"a"}, []string{"a", "b", "c"}, false, true, false); err == nil {
		t.Error("compare must surface a write error from the second tail loop")
	}
}

// TestCompareSuppressAllNoOutput verifies that suppressing every column writes
// nothing, so even a failing writer is never touched.
func TestCompareSuppressAllNoOutput(t *testing.T) {
	t.Parallel()
	io := errIO(&failingWriter{err: errors.New("must not be called")})
	c := New()
	if err := c.compare(io, []string{"a"}, []string{"a", "b"}, false, false, false); err != nil {
		t.Errorf("suppressing all columns must not write or error, got %v", err)
	}
}
