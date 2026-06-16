package cut_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/cut"
	"github.com/nao1215/mimixbox/internal/command"
)

func runStdin(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := cut.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunCut(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"field 1 default tab", "a\tb\tc\n", []string{"-f", "1"}, "a\n"},
		{"field 2 comma delim", "a,b,c\n", []string{"-f", "2", "-d", ","}, "b\n"},
		{"fields 1,3 comma", "a,b,c\n", []string{"-f", "1,3", "-d", ","}, "a,c\n"},
		{"fields 2- comma", "a,b,c\n", []string{"-f", "2-", "-d", ","}, "b,c\n"},
		{"chars 1-3", "abcdef\n", []string{"-c", "1-3"}, "abc\n"},
		{"byte 1", "abc\n", []string{"-b", "1"}, "a\n"},
		{
			"only delimited suppresses",
			"nodelim\na,b\n",
			[]string{"-f", "1", "-d", ",", "-s"},
			"a\n",
		},
		{
			"output delimiter",
			"a,b,c\n",
			[]string{"-f", "1,3", "-d", ",", "--output-delimiter=:"},
			"a:c\n",
		},
		{
			"no delimiter line passes through",
			"nodelim\na,b\n",
			[]string{"-f", "1", "-d", ","},
			"nodelim\na\n",
		},
		{
			"long flags",
			"a,b,c\n",
			[]string{"--fields=2", "--delimiter=,"},
			"b\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := runStdin(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRunNoListSpecified(t *testing.T) {
	t.Parallel()
	out, errOut, err := runStdin(t, "a,b\n")
	if err == nil {
		t.Fatal("expected error when no list is specified")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	want := "cut: you must specify a list of bytes, characters, or fields"
	if !strings.Contains(errOut, want) {
		t.Errorf("stderr = %q, want to contain %q", errOut, want)
	}
}

func TestRunMultipleListsError(t *testing.T) {
	t.Parallel()
	_, errOut, err := runStdin(t, "a,b\n", "-b", "1", "-f", "1")
	if err == nil {
		t.Fatal("expected error when more than one list is specified")
	}
	if !strings.Contains(errOut, "only one type of list may be specified") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunInvalidList(t *testing.T) {
	t.Parallel()
	_, errOut, err := runStdin(t, "abc\n", "-c", "0")
	if err == nil {
		t.Fatal("expected error for position 0")
	}
	if !strings.Contains(errOut, "numbered from 1") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestRangeMergingAndBounds drives mergeRanges and the out-of-bounds clamping
// in cutBytes/cutRunes via observable output.
func TestRangeMergingAndBounds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"overlapping ranges merge", "abcdef\n", []string{"-c", "1-3,2-4"}, "abcd\n"},
		{"adjacent ranges merge", "abcdef\n", []string{"-c", "1-2,3-4"}, "abcd\n"},
		{"open range absorbs later", "abcdef\n", []string{"-c", "2-,4-5"}, "bcdef\n"},
		{"open range earlier wins", "abcdef\n", []string{"-c", "1-,3"}, "abcdef\n"},
		{"byte range past end clamps", "abc\n", []string{"-b", "2-10"}, "bc\n"},
		{"byte lo past end yields empty", "abc\n", []string{"-b", "5-7"}, "\n"},
		{"char range past end clamps", "abc\n", []string{"-c", "2-10"}, "bc\n"},
		{"char lo past end yields empty", "abc\n", []string{"-c", "5-7"}, "\n"},
		{"multibyte runes", "héllo\n", []string{"-c", "1-3"}, "hél\n"},
		{"open from start", "abcdef\n", []string{"-c", "-3"}, "abc\n"},
		{"duplicate single positions", "abcdef\n", []string{"-c", "2,2,2"}, "b\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := runStdin(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestRunDecreasingRange covers the invalid-decreasing-range branch of parseRange.
func TestRunDecreasingRange(t *testing.T) {
	t.Parallel()
	_, errOut, err := runStdin(t, "abc\n", "-c", "5-2")
	if err == nil {
		t.Fatal("expected error for decreasing range")
	}
	if !strings.Contains(errOut, "decreasing range") {
		t.Errorf("stderr = %q, want decreasing range", errOut)
	}
}

// TestRunEmptyListItem covers the empty-item branch of parseRanges.
func TestRunEmptyListItem(t *testing.T) {
	t.Parallel()
	_, errOut, err := runStdin(t, "abc\n", "-c", "1,,3")
	if err == nil {
		t.Fatal("expected error for empty list item")
	}
	if !strings.Contains(errOut, "invalid byte, character or field list") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestRunNonNumericRange covers the parsePos non-numeric branch.
func TestRunNonNumericRange(t *testing.T) {
	t.Parallel()
	_, errOut, err := runStdin(t, "abc\n", "-c", "a-b")
	if err == nil {
		t.Fatal("expected error for non-numeric range")
	}
	if !strings.Contains(errOut, "invalid byte, character or field list") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestRunMissingFile exercises the open-error path in run() (keep()).
func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := cut.New().Run(context.Background(), io, []string{"-c", "1", "/no/such/cut/file"})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(errBuf.String(), "cut:") {
		t.Errorf("stderr = %q, want cut: prefix", errBuf.String())
	}
}

// TestSynopsis ensures the one-line description is reported.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	if s := cut.New().Synopsis(); s == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := runStdin(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: cut") {
		t.Errorf("--help out = %q", out)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q:\n%s", want, out)
		}
	}

	out, _, err = runStdin(t, "", "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "cut (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}
