package cmp

import (
	"errors"
	"strings"
	"testing"
)

// failingReader always returns its configured error, modeling a stream that
// fails partway through with something other than io.EOF.
type failingReader struct{ err error }

func (f failingReader) Read([]byte) (int, error) { return 0, f.err }

// TestCompareReadError verifies compare surfaces a non-EOF read error from
// either side rather than misreporting equality or a difference.
func TestCompareReadError(t *testing.T) {
	t.Parallel()
	boom := errors.New("boom")

	if _, err := compare(failingReader{err: boom}, strings.NewReader("x")); !errors.Is(err, boom) {
		t.Errorf("error from first reader = %v, want boom", err)
	}
	if _, err := compare(strings.NewReader("x"), failingReader{err: boom}); !errors.Is(err, boom) {
		t.Errorf("error from second reader = %v, want boom", err)
	}
}

// TestCompareEqualAndDiff exercises the pure comparison directly: equal streams,
// a concrete byte difference (with its line number), and prefix relationships.
func TestCompareEqualAndDiff(t *testing.T) {
	t.Parallel()

	res, err := compare(strings.NewReader("ab\ncd\n"), strings.NewReader("ab\ncd\n"))
	if err != nil || !res.equal {
		t.Errorf("equal streams: res=%+v err=%v", res, err)
	}

	// First difference at byte 5 (the 'c'/'C'), on line 2.
	res, err = compare(strings.NewReader("ab\ncd"), strings.NewReader("ab\nCd"))
	if err != nil {
		t.Fatalf("diff err = %v", err)
	}
	if res.equal || res.eofOn != 0 || res.byteOffset != 4 || res.line != 2 || res.a != 'c' || res.b != 'C' {
		t.Errorf("diff res = %+v, want byte 4 line 2 c/C", res)
	}

	res, _ = compare(strings.NewReader("abc"), strings.NewReader("abcd"))
	if res.eofOn != 1 {
		t.Errorf("first-prefix res = %+v, want eofOn=1", res)
	}
	res, _ = compare(strings.NewReader("abcd"), strings.NewReader("abc"))
	if res.eofOn != 2 {
		t.Errorf("second-prefix res = %+v, want eofOn=2", res)
	}
}
