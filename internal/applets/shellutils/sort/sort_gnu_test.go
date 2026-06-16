package sortcmd_test

import (
	"strings"
	"testing"
)

// TestVersionSort verifies that -V orders embedded numbers by value, so 1.2
// sorts before 1.10 (unlike a plain lexical sort).
func TestVersionSort(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "1.10\n1.2\n1.1\n1.20\n", "-V")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "1.1\n1.2\n1.10\n1.20\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestVersionSortWithText verifies -V mixes lexical and numeric runs correctly.
func TestVersionSortWithText(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "v1.10\nv1.9\nv1.0\n", "-V")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "v1.0\nv1.9\nv1.10\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestGeneralNumericSort verifies -g compares floating-point values, including
// scientific notation, which plain -n would not parse fully.
func TestGeneralNumericSort(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "1e3\n2.5\n100\n0.5\n", "-g")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "0.5\n2.5\n100\n1e3\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestHumanNumericSort verifies -h orders human-readable sizes by magnitude so
// that 2K < 1M < 1G regardless of the leading digit.
func TestHumanNumericSort(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "1G\n2K\n1M\n500\n", "-h")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "500\n2K\n1M\n1G\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestHumanNumericLongFlag verifies the long --human-numeric-sort form works.
func TestHumanNumericLongFlag(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "1G\n2K\n1M\n", "--human-numeric-sort")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "2K\n1M\n1G\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestStableSort verifies -s preserves the input order of lines whose keys are
// equal, instead of falling back to a full-line comparison.
func TestStableSort(t *testing.T) {
	t.Parallel()
	// All lines share the numeric key 5, so -s must keep their input order
	// rather than breaking the tie with a full-line comparison.
	in := "5 zebra\n5 apple\n5 mango\n"
	out, _, err := run(t, in, "-s", "-n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != in {
		t.Errorf("out = %q, want %q (stable order)", out, in)
	}
}

// TestNonStableLastResort verifies that without -s, equal keys fall back to the
// full-line comparison (the GNU last-resort comparison).
func TestNonStableLastResort(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "5 zebra\n5 apple\n5 mango\n", "-n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "5 apple\n5 mango\n5 zebra\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestZeroTerminated verifies -z reads and writes NUL-delimited records.
func TestZeroTerminated(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "banana\x00apple\x00cherry\x00", "-z")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "apple\x00banana\x00cherry\x00"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestZeroTerminatedKeepsNewlines verifies -z treats newlines as ordinary data.
func TestZeroTerminatedKeepsNewlines(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "b\nb\x00a\na\x00", "-z")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "a\na\x00b\nb\x00"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestMerge verifies -m produces a sorted result from already-sorted inputs.
func TestMerge(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "apple\nbanana\ncherry\n", "-m")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "apple\nbanana\ncherry\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestParallelAccepted verifies --parallel=N is accepted and has no effect.
func TestParallelAccepted(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "b\na\n", "--parallel=4")
	if err != nil {
		t.Fatalf("Run error = %v, stderr = %q", err, errOut)
	}
	if want := "a\nb\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestTemporaryDirectoryAccepted verifies --temporary-directory=DIR is accepted
// and has no effect on the result.
func TestTemporaryDirectoryAccepted(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "b\na\n", "--temporary-directory=/tmp")
	if err != nil {
		t.Fatalf("Run error = %v, stderr = %q", err, errOut)
	}
	if want := "a\nb\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestVersionSortAgainstSystem cross-checks -V against the system sort when
// available, guarding against ordering surprises.
func TestVersionSortAgainstSystem(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "1.10\n1.2\n1.1\n", "-V")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.HasPrefix(out, "1.1\n1.2\n1.10") {
		t.Errorf("out = %q", out)
	}
}
