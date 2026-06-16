package dd_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/dd"
	"github.com/nao1215/mimixbox/internal/command"
)

// run executes dd with the given stdin string and arguments, returning the
// captured stdout, stderr, and the error.
func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{
		In:  strings.NewReader(stdin),
		Out: &out,
		Err: &errBuf,
	}
	err := dd.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNew(t *testing.T) {
	c := dd.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Name() != "dd" {
		t.Errorf("Name() = %q, want %q", c.Name(), "dd")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestRun_StdinToStdout(t *testing.T) {
	input := "hello, world\nsecond line\n"
	out, errOut, err := run(t, input)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != input {
		t.Errorf("stdout = %q, want %q", out, input)
	}
	if !strings.Contains(errOut, "records in") || !strings.Contains(errOut, "records out") {
		t.Errorf("stderr missing records summary: %q", errOut)
	}
}

func TestRun_BsCount(t *testing.T) {
	input := "abcdefghij"
	out, errOut, err := run(t, input, "bs=1", "count=5")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "abcde" {
		t.Errorf("stdout = %q, want %q", out, "abcde")
	}
	if !strings.Contains(errOut, "5+0 records in") {
		t.Errorf("stderr = %q, want to contain %q", errOut, "5+0 records in")
	}
}

func TestRun_FileToFile(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.txt")
	outPath := filepath.Join(dir, "out.txt")
	content := []byte("the quick brown fox\njumps over the lazy dog\n")
	if err := os.WriteFile(inPath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := run(t, "", "if="+inPath, "of="+outPath)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("output file = %q, want %q", got, content)
	}
}

func TestRun_ConvUcase(t *testing.T) {
	out, _, err := run(t, "Hello World 123", "conv=ucase")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "HELLO WORLD 123" {
		t.Errorf("stdout = %q, want %q", out, "HELLO WORLD 123")
	}
}

func TestRun_ConvLcase(t *testing.T) {
	out, _, err := run(t, "Hello World 123", "conv=lcase")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "hello world 123" {
		t.Errorf("stdout = %q, want %q", out, "hello world 123")
	}
}

func TestRun_StatusNoneSuppressesSummary(t *testing.T) {
	out, errOut, err := run(t, "data", "status=none")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "data" {
		t.Errorf("stdout = %q, want %q", out, "data")
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty (status=none)", errOut)
	}
}

func TestRun_InvalidOperand(t *testing.T) {
	_, _, err := run(t, "data", "bogus")
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil for invalid operand")
	}
}

func TestParseSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"512", 512, false},
		{"1k", 1024, false},
		{"1K", 1024, false},
		{"2b", 1024, false},
		{"1M", 1024 * 1024, false},
		{"1c", 1, false},
		{"2w", 4, false},
		{"1G", 1024 * 1024 * 1024, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
		{"-5", 0, true},
	}
	for _, tt := range tests {
		got, err := dd.ParseSize(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseSize(%q) error = nil, want error", tt.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseSize(%q) error = %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseSize(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

// TestRun_Skip skips the first ibs-sized block of input.
func TestRun_Skip(t *testing.T) {
	out, _, err := run(t, "0123456789", "bs=2", "skip=1")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "23456789" {
		t.Errorf("stdout = %q, want %q", out, "23456789")
	}
}

// TestRun_Seek pads the output with NUL bytes before writing (stdout path).
func TestRun_Seek(t *testing.T) {
	out, _, err := run(t, "abc", "bs=1", "seek=2")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	want := "\x00\x00abc"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
}

// TestRun_SeekFile seeks within an output file.
func TestRun_SeekFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.bin")
	if _, _, err := run(t, "XYZ", "bs=1", "seek=3", "of="+outPath); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("\x00\x00\x00XYZ")
	if !bytes.Equal(got, want) {
		t.Errorf("file = %q, want %q", got, want)
	}
}

// TestRun_ConvNotrunc keeps the tail of a pre-existing output file instead of
// truncating it.
func TestRun_ConvNotrunc(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(outPath, []byte("OLDDATA-TAIL"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, "NEW", "bs=1", "conv=notrunc", "of="+outPath); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	// "NEW" overwrites the first three bytes; the tail survives.
	if want := "NEWDATA-TAIL"; string(got) != want {
		t.Errorf("file = %q, want %q", got, want)
	}
}

// TestRun_ConvSync pads a short final block to ibs with NUL.
func TestRun_ConvSync(t *testing.T) {
	out, errOut, err := run(t, "abcde", "ibs=4", "obs=4", "conv=sync")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	// First full block "abcd" plus partial "e" padded to 4 bytes.
	want := "abcde\x00\x00\x00"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
	if !strings.Contains(errOut, "1+1 records in") {
		t.Errorf("stderr = %q, want to contain %q", errOut, "1+1 records in")
	}
}

// TestRun_StatusNoxferSuppressesByteLine drops the trailing "bytes copied" line.
func TestRun_StatusNoxfer(t *testing.T) {
	_, errOut, err := run(t, "data", "status=noxfer")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(errOut, "records out") {
		t.Errorf("stderr = %q, want records summary", errOut)
	}
	if strings.Contains(errOut, "bytes copied") {
		t.Errorf("stderr = %q, want no bytes-copied line for status=noxfer", errOut)
	}
}

// TestRun_StatusProgress behaves like the default summary.
func TestRun_StatusProgress(t *testing.T) {
	_, errOut, err := run(t, "data", "status=progress")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(errOut, "bytes copied") {
		t.Errorf("stderr = %q, want bytes-copied line for status=progress", errOut)
	}
}

// TestRun_OpenInputError reports a missing input file.
func TestRun_OpenInputError(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.bin")
	_, errOut, err := run(t, "", "if="+missing)
	if err == nil {
		t.Fatal("Run() error = nil, want error for missing input")
	}
	if !strings.Contains(errOut, "dd:") {
		t.Errorf("stderr = %q, want dd: prefix", errOut)
	}
}

// TestRun_OpenOutputError reports an output path that cannot be created.
func TestRun_OpenOutputError(t *testing.T) {
	dir := t.TempDir()
	// A path whose parent component is a file, not a directory.
	notDir := filepath.Join(dir, "file")
	if err := os.WriteFile(notDir, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(notDir, "child")
	_, errOut, err := run(t, "data", "of="+badPath)
	if err == nil {
		t.Fatal("Run() error = nil, want error for bad output path")
	}
	if !strings.Contains(errOut, "dd:") {
		t.Errorf("stderr = %q, want dd: prefix", errOut)
	}
}

// TestRun_InvalidConv exercises the conv= error branch in parseArgs.
func TestRun_InvalidConv(t *testing.T) {
	_, errOut, err := run(t, "data", "conv=bogus")
	if err == nil {
		t.Fatal("Run() error = nil, want error for unknown conv")
	}
	if !strings.Contains(errOut, "unknown conversion") {
		t.Errorf("stderr = %q, want unknown conversion", errOut)
	}
}

// TestRun_ConvLcaseUcaseConflict rejects combining lcase and ucase.
func TestRun_ConvLcaseUcaseConflict(t *testing.T) {
	_, errOut, err := run(t, "data", "conv=lcase,ucase")
	if err == nil {
		t.Fatal("Run() error = nil, want error for lcase+ucase")
	}
	if !strings.Contains(errOut, "cannot combine") {
		t.Errorf("stderr = %q, want cannot combine", errOut)
	}
}

// TestRun_InvalidStatus exercises the status= error branch.
func TestRun_InvalidStatus(t *testing.T) {
	_, errOut, err := run(t, "data", "status=bogus")
	if err == nil {
		t.Fatal("Run() error = nil, want error for unknown status")
	}
	if !strings.Contains(errOut, "unknown status") {
		t.Errorf("stderr = %q, want unknown status", errOut)
	}
}

// TestRun_InvalidSizes drives parseSize error branches via every operand.
func TestRun_InvalidSizes(t *testing.T) {
	t.Parallel()
	for _, op := range []string{"bs", "ibs", "obs", "count", "skip", "seek"} {
		op := op
		t.Run(op, func(t *testing.T) {
			t.Parallel()
			_, errOut, err := run(t, "data", op+"=notanumber")
			if err == nil {
				t.Fatalf("Run() error = nil, want error for %s=notanumber", op)
			}
			if !strings.Contains(errOut, "invalid "+op) {
				t.Errorf("stderr = %q, want to contain %q", errOut, "invalid "+op)
			}
		})
	}
}

// TestRun_IbsObsRespectedWhenNoBs ensures ibs=/obs= set independently when bs=
// is absent.
func TestRun_IbsObs(t *testing.T) {
	out, errOut, err := run(t, "aabbccdd", "ibs=2", "obs=4")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "aabbccdd" {
		t.Errorf("stdout = %q, want %q", out, "aabbccdd")
	}
	// 4 full input blocks of 2 bytes each. Output is written one input block
	// at a time, so each 2-byte write is a partial obs(=4) block: 0+4 out.
	if !strings.Contains(errOut, "4+0 records in") {
		t.Errorf("stderr = %q, want 4+0 records in", errOut)
	}
	if !strings.Contains(errOut, "0+4 records out") {
		t.Errorf("stderr = %q, want 0+4 records out", errOut)
	}
}

// TestRun_BsOverridesIbsObs ensures bs= wins over a later ibs=/obs=.
func TestRun_BsOverridesIbsObs(t *testing.T) {
	out, _, err := run(t, "abcd", "bs=2", "ibs=99", "obs=99")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "abcd" {
		t.Errorf("stdout = %q, want %q", out, "abcd")
	}
}

// TestHelpSections asserts `dd --help` renders structured help.
func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := dd.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Usage: dd", "Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q: %q", want, out.String())
		}
	}
}
