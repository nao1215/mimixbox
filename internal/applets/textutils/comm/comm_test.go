package comm_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/comm"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := comm.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestThreeColumns(t *testing.T) {
	t.Parallel()
	a := writeFile(t, "apple\nbanana\ncherry\n")
	b := writeFile(t, "banana\ncherry\ndate\n")
	out, _, err := run(t, a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "apple\n\t\tbanana\n\t\tcherry\n\tdate\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestSuppressColumns(t *testing.T) {
	t.Parallel()
	a := writeFile(t, "apple\nbanana\ncherry\n")
	b := writeFile(t, "banana\ncherry\ndate\n")
	// -1 -2 leaves only the common column, with no indentation.
	out, _, err := run(t, "-1", "-2", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "banana\ncherry\n" {
		t.Errorf("out = %q", out)
	}
}

func TestOnlyUniqueToFirst(t *testing.T) {
	t.Parallel()
	a := writeFile(t, "apple\nbanana\n")
	b := writeFile(t, "banana\n")
	out, _, err := run(t, "-2", "-3", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "apple\n" {
		t.Errorf("out = %q", out)
	}
}

// TestOutputDelimiter verifies --output-delimiter=STR replaces the default tab
// column separators with STR, including the per-column padding.
func TestOutputDelimiter(t *testing.T) {
	t.Parallel()
	a := writeFile(t, "apple\nbanana\ncherry\n")
	b := writeFile(t, "banana\ncherry\ndate\n")
	out, _, err := run(t, "--output-delimiter=,", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// col1 has no pad; col2 one delimiter; col3 two delimiters.
	want := "apple\n,,banana\n,,cherry\n,date\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestEmptyOutputDelimiter verifies an explicitly empty --output-delimiter is
// rejected (GNU treats it as an error).
func TestEmptyOutputDelimiter(t *testing.T) {
	t.Parallel()
	a := writeFile(t, "a\n")
	b := writeFile(t, "b\n")
	_, errOut, err := run(t, "--output-delimiter=", a, b)
	if err == nil {
		t.Fatal("expected error for empty --output-delimiter")
	}
	if !strings.Contains(errOut, "empty --output-delimiter") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestZeroTerminated verifies -z reads NUL-separated input records and writes
// NUL-terminated output records.
func TestZeroTerminated(t *testing.T) {
	t.Parallel()
	a := writeFile(t, "apple\x00banana\x00")
	b := writeFile(t, "banana\x00date\x00")
	out, _, err := run(t, "-z", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// Records terminated by NUL; columns still separated by tab padding.
	want := "apple\x00\t\tbanana\x00\tdate\x00"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestCheckOrderUnsorted verifies --check-order reports an out-of-order input on
// stderr and fails, naming the offending file.
func TestCheckOrderUnsorted(t *testing.T) {
	t.Parallel()
	// FILE1 is unsorted (cherry before banana).
	a := writeFile(t, "cherry\nbanana\n")
	b := writeFile(t, "apple\nbanana\n")
	_, errOut, err := run(t, "--check-order", a, b)
	if err == nil {
		t.Fatal("expected failure for unsorted input")
	}
	if !strings.Contains(errOut, "comm: file 1 is not in sorted order") {
		t.Errorf("stderr = %q, want file-1-unsorted message", errOut)
	}
}

// TestCheckOrderSecondUnsorted verifies the file-2 disorder branch.
func TestCheckOrderSecondUnsorted(t *testing.T) {
	t.Parallel()
	a := writeFile(t, "apple\nbanana\n")
	b := writeFile(t, "date\nbanana\n")
	_, errOut, err := run(t, "--check-order", a, b)
	if err == nil {
		t.Fatal("expected failure for unsorted second input")
	}
	if !strings.Contains(errOut, "comm: file 2 is not in sorted order") {
		t.Errorf("stderr = %q, want file-2-unsorted message", errOut)
	}
}

// TestCheckOrderSorted verifies --check-order succeeds on properly sorted input.
func TestCheckOrderSorted(t *testing.T) {
	t.Parallel()
	a := writeFile(t, "apple\nbanana\n")
	b := writeFile(t, "banana\ncherry\n")
	out, _, err := run(t, "--check-order", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "apple\n\t\tbanana\n\tcherry\n" {
		t.Errorf("out = %q", out)
	}
}

func TestWrongOperandCount(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "onlyone")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "two file operands") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	b := writeFile(t, "x\n")
	_, errOut, err := run(t, "/no/such/file", b)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "comm: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := comm.New()
	if c.Name() != "comm" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestHelpSections verifies that --help renders both the Examples and the
// Exit status sections supplied through WithHelp.
func TestHelpSections(t *testing.T) {
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help missing %q section:\n%s", want, out)
		}
	}
}
