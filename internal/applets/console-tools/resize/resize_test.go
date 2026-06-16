package resize

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// fakeSize installs a deterministic winsize for the duration of a test.
func fakeSize(t *testing.T, rows, cols uint16, err error) {
	t.Helper()
	orig := winsize
	winsize = func() (uint16, uint16, error) { return rows, cols, err }
	t.Cleanup(func() { winsize = orig })
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if got := c.Name(); got != "resize" {
		t.Errorf("Name() = %q, want %q", got, "resize")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestRunDefaultShOutput(t *testing.T) {
	fakeSize(t, 24, 80, nil)
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "COLUMNS=80;\nLINES=24;\nexport COLUMNS LINES;\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunCshOutput(t *testing.T) {
	fakeSize(t, 30, 100, nil)
	out, _, err := run(t, "-c")
	if err != nil {
		t.Fatalf("Run -c error = %v", err)
	}
	want := "set noglob;\nsetenv COLUMNS '100';\nsetenv LINES '30';\nunset noglob;\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunShFlagBeatsCsh(t *testing.T) {
	fakeSize(t, 24, 80, nil)
	out, _, err := run(t, "-c", "-u")
	if err != nil {
		t.Fatalf("Run -c -u error = %v", err)
	}
	if !strings.Contains(out, "export COLUMNS LINES;") {
		t.Errorf("with -u, expected sh-style output, got %q", out)
	}
}

func TestRunError(t *testing.T) {
	fakeSize(t, 0, 0, errFake{})
	_, errOut, err := run(t)
	if err == nil {
		t.Error("expected error when winsize fails")
	}
	if !strings.Contains(errOut, "resize:") {
		t.Errorf("stderr = %q, want resize: prefix", errOut)
	}
}

func TestFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		rows uint16
		cols uint16
		csh  bool
		want string
	}{
		{"sh", 24, 80, false, "COLUMNS=80;\nLINES=24;\nexport COLUMNS LINES;\n"},
		{"csh", 50, 132, true, "set noglob;\nsetenv COLUMNS '132';\nsetenv LINES '50';\nunset noglob;\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := format(tt.rows, tt.cols, tt.csh); got != tt.want {
				t.Errorf("format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	for _, want := range []string{"Usage: resize", "Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q\n%s", want, out)
		}
	}
}

// TestRealWinsize exercises the default winsize implementation. The test runner
// usually has no controlling terminal, so this normally returns an error; when
// it does run under a tty it returns positive dimensions. Either way the ioctl
// path is executed.
func TestRealWinsize(t *testing.T) {
	rows, cols, err := winsize()
	if err == nil && (rows == 0 || cols == 0) {
		t.Errorf("winsize returned no error but zero size: rows=%d cols=%d", rows, cols)
	}
}

// errFake is a stand-in error for the winsize failure path.
type errFake struct{}

func (errFake) Error() string { return "no tty" }
