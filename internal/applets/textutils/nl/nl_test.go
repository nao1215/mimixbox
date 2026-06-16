package nl_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/nl"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := nl.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"default skips blank", "a\n\nb\n", nil, "     1\ta\n       \n     2\tb\n"},
		{"number all", "a\n\nb\n", []string{"-b", "a"}, "     1\ta\n     2\t\n     3\tb\n"},
		{"number none", "a\nb\n", []string{"-b", "n"}, "       a\n       b\n"},
		{"separator and width", "a\n", []string{"-s", ": ", "-w", "3"}, "  1: a\n"},
		{"zero format", "a\n", []string{"-n", "rz", "-w", "3"}, "001\ta\n"},
		{"start and increment", "a\nb\n", []string{"-v", "5", "-i", "5", "-b", "a"}, "     5\ta\n    10\tb\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRunInvalidStyle(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "a\n", "-b", "z")
	if err == nil {
		t.Fatal("expected error for invalid style")
	}
	if !strings.Contains(errOut, "invalid body numbering style") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameAndSynopsis(t *testing.T) {
	t.Parallel()
	c := nl.New()
	if c.Name() != "nl" {
		t.Errorf("Name() = %q, want %q", c.Name(), "nl")
	}
	want := "Write each FILE to standard output with line numbers added"
	if c.Synopsis() != want {
		t.Errorf("Synopsis() = %q, want %q", c.Synopsis(), want)
	}
}

func TestRunInvalidFormat(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "a\n", "-n", "zz")
	if err == nil {
		t.Fatal("expected error for invalid number format")
	}
	if !strings.Contains(errOut, "invalid line numbering format") {
		t.Errorf("stderr = %q, want it to mention invalid format", errOut)
	}
}

func TestRunLeftFormat(t *testing.T) {
	t.Parallel()
	// ln (left-justified) numbers padded on the right within the width field.
	out, _, err := run(t, "a\n", "-n", "ln", "-w", "3", "-b", "a")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "1  \ta\n" {
		t.Errorf("out = %q, want %q", out, "1  \ta\n")
	}
}

func TestRunMissingFileKeepsError(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "nl:") {
		t.Errorf("stderr = %q, want an nl error", errOut)
	}
}
