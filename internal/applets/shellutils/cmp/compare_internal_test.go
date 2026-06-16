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

	if _, err := compare(failingReader{err: boom}, strings.NewReader("x"), compareOptions{}); !errors.Is(err, boom) {
		t.Errorf("error from first reader = %v, want boom", err)
	}
	if _, err := compare(strings.NewReader("x"), failingReader{err: boom}, compareOptions{}); !errors.Is(err, boom) {
		t.Errorf("error from second reader = %v, want boom", err)
	}
}

// TestCompareEqualAndDiff exercises the pure comparison directly: equal streams,
// a concrete byte difference (with its line number), and prefix relationships.
func TestCompareEqualAndDiff(t *testing.T) {
	t.Parallel()

	res, err := compare(strings.NewReader("ab\ncd\n"), strings.NewReader("ab\ncd\n"), compareOptions{})
	if err != nil || !res.equal {
		t.Errorf("equal streams: res=%+v err=%v", res, err)
	}

	// First difference at byte 5 (the 'c'/'C'), on line 2.
	res, err = compare(strings.NewReader("ab\ncd"), strings.NewReader("ab\nCd"), compareOptions{})
	if err != nil {
		t.Fatalf("diff err = %v", err)
	}
	if res.equal || res.eofOn != 0 || res.byteOffset != 4 || res.line != 2 || res.a != 'c' || res.b != 'C' {
		t.Errorf("diff res = %+v, want byte 4 line 2 c/C", res)
	}

	res, _ = compare(strings.NewReader("abc"), strings.NewReader("abcd"), compareOptions{})
	if res.eofOn != 1 {
		t.Errorf("first-prefix res = %+v, want eofOn=1", res)
	}
	res, _ = compare(strings.NewReader("abcd"), strings.NewReader("abc"), compareOptions{})
	if res.eofOn != 2 {
		t.Errorf("second-prefix res = %+v, want eofOn=2", res)
	}
}

// TestCompareLimit verifies --bytes (limit) stops the comparison early: a
// difference past the limit is not seen, so the streams compare equal.
func TestCompareLimit(t *testing.T) {
	t.Parallel()

	// Differ at byte 3, but limit to 2 bytes -> equal.
	res, err := compare(strings.NewReader("abX"), strings.NewReader("abY"), compareOptions{limit: 2})
	if err != nil || !res.equal {
		t.Errorf("limited compare: res=%+v err=%v, want equal", res, err)
	}

	// Limit equal to the difference offset still catches it (offset 3 with limit 3).
	res, err = compare(strings.NewReader("abX"), strings.NewReader("abY"), compareOptions{limit: 3})
	if err != nil {
		t.Fatalf("limit err = %v", err)
	}
	if res.equal || res.byteOffset != 3 {
		t.Errorf("limit-at-diff res = %+v, want byte 3 diff", res)
	}
}

// TestCompareIgnoreInitial verifies --ignore-initial skips leading bytes of each
// stream, and that offsets are counted from the first compared byte.
func TestCompareIgnoreInitial(t *testing.T) {
	t.Parallel()

	// Skip 3 of each; the remaining "abc" vs "abc" are equal.
	res, err := compare(strings.NewReader("XYZabc"), strings.NewReader("123abc"), compareOptions{skip1: 3, skip2: 3})
	if err != nil || !res.equal {
		t.Errorf("skip N: res=%+v err=%v, want equal", res, err)
	}

	// Asymmetric skip N:M. Skip 1 of file1 and 3 of file2; remaining "bcd" vs
	// "bcd" equal. Verifies the two counts are independent.
	res, err = compare(strings.NewReader("Abcd"), strings.NewReader("XYZbcd"), compareOptions{skip1: 1, skip2: 3})
	if err != nil || !res.equal {
		t.Errorf("skip N:M: res=%+v err=%v, want equal", res, err)
	}

	// After skipping, the first compared byte is offset 1.
	res, err = compare(strings.NewReader("XXa"), strings.NewReader("XXb"), compareOptions{skip1: 2, skip2: 2})
	if err != nil {
		t.Fatalf("skip-diff err = %v", err)
	}
	if res.byteOffset != 1 || res.a != 'a' || res.b != 'b' {
		t.Errorf("skip-diff res = %+v, want byte 1 a/b", res)
	}
}
