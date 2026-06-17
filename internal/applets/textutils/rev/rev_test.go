package rev_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/rev"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := rev.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestLongLineBeyondDefaultScannerCap(t *testing.T) {
	t.Parallel()
	// A 2 MiB single line exceeds the old fixed scanner cap; rev must handle it
	// like GNU rev instead of failing with "token too long" (issue #950).
	line := strings.Repeat("a", 2*1024*1024)
	out, errOut, err := run(t, line+"\n")
	if err != nil {
		t.Fatalf("rev error = %v (stderr: %s)", err, errOut)
	}
	if want := line + "\n"; out != want {
		t.Errorf("rev of a 2 MiB line: got %d bytes, want %d", len(out), len(want))
	}
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"single line", "abc\n", nil, "cba\n"},
		{"multiple lines", "abc\ndef\n", nil, "cba\nfed\n"},
		{"no trailing newline", "abc", nil, "cba\n"},
		{"empty line", "\n", nil, "\n"},
		{"utf8 aware", "あいう\n", nil, "ういあ\n"},
		{"explicit stdin", "hello\n", []string{"-"}, "olleh\n"},
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

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "rev: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := rev.New()
	if c.Name() != "rev" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	for _, want := range []string{"Usage: rev", "Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q\n%s", want, out)
		}
	}
}
