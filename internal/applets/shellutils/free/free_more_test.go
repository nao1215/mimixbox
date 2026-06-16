package free

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// TestRunMeminfoSourceError covers the branch where /proc/meminfo cannot be
// opened: Run must report a failure rather than print a report.
func TestRunMeminfoSourceError(t *testing.T) {
	orig := meminfoSource
	meminfoSource = func() (io.Reader, error) { return nil, errors.New("boom") }
	t.Cleanup(func() { meminfoSource = orig })

	out, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected an error when meminfo cannot be read")
	}
	if out != "" {
		t.Errorf("no report should be printed on error, got %q", out)
	}
	_ = errOut
}

// errReader fails partway through, so bufio.Scanner surfaces the error via
// sc.Err() and parseMeminfo returns it.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

// TestRunMeminfoReadError covers the parseMeminfo error branch.
func TestRunMeminfoReadError(t *testing.T) {
	orig := meminfoSource
	meminfoSource = func() (io.Reader, error) { return errReader{}, nil }
	t.Cleanup(func() { meminfoSource = orig })
	if _, _, err := run(t); err == nil {
		t.Fatal("expected an error when meminfo cannot be read")
	}
}

// TestGibibytes covers chooseUnit's gibibyte branch and render at that scale.
func TestGibibytes(t *testing.T) {
	stubMeminfo(t, sampleMeminfo)
	out, _, err := run(t, "-g")
	if err != nil {
		t.Fatalf("Run -g error = %v", err)
	}
	// 8000000 KiB / (1024*1024) rounds down to 7 GiB.
	if !strings.Contains(out, "Mem:") || !strings.Contains(out, "7") {
		t.Errorf("gibibyte output unexpected: %q", out)
	}
}
