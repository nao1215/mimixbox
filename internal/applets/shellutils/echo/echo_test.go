package echo_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/echo"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := echo.New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"plain", []string{"hello", "world"}, "hello world\n"},
		{"no args", nil, "\n"},
		{"no newline", []string{"-n", "hi"}, "hi"},
		{"unknown flag is literal", []string{"-x", "y"}, "-x y\n"},
		{"help not first is literal", []string{"foo", "--help"}, "foo --help\n"},
		{"version not first is literal", []string{"foo", "--version"}, "foo --version\n"},
		{"escapes off by default", []string{`a\tb`}, "a\\tb\n"},
		{"escapes on", []string{"-e", `a\tb`}, "a\tb\n"},
		{"combined flags", []string{"-ne", `a\nb`}, "a\nb"},
		{"escape c stops output", []string{"-e", `ab\ccd`}, "ab"},
		{"hex escape", []string{"-e", `\x41`}, "A\n"},
		{"octal escape", []string{"-e", `\0101`}, "A\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := run(t, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestEscapeExpansion drives every backslash escape branch of expandEscapes,
// including octal/hex edge cases and malformed escapes.
func TestEscapeExpansion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"alert", `\a`, "\a\n"},
		{"backspace", `\b`, "\b\n"},
		{"formfeed", `\f`, "\f\n"},
		{"carriage return", `\r`, "\r\n"},
		{"vertical tab", `\v`, "\v\n"},
		{"literal backslash", `\\`, "\\\n"},
		{"newline", `\n`, "\n\n"},
		{"unknown escape kept literal", `\q`, "\\q\n"},
		{"trailing backslash kept", `end\`, "end\\\n"},
		{"hex two digits", `\x4a`, "J\n"},
		{"hex one digit", `\x9z`, "\tz\n"},
		{"hex no digits kept literal", `\xz`, "\\xz\n"},
		{"octal full", `\0101`, "A\n"},
		{"octal short", `\007`, "\a\n"},
		{"octal zero only", `\0`, "\x00\n"},
		{"text around escape", `a\tb\tc`, "a\tb\tc\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := run(t, "-e", tt.in)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestEFlagThenEDisablesEscapes verifies -E after -e turns interpretation off.
func TestEFlagThenEDisablesEscapes(t *testing.T) {
	t.Parallel()
	out, err := run(t, "-e", "-E", `a\tb`)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "a\\tb\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestSynopsis ensures the one-line description is reported.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	if s := echo.New().Synopsis(); s == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestHelpAsFirstArg verifies that a leading --help prints usage rather than
// echoing the literal text, matching GNU's standalone echo.
func TestHelpAsFirstArg(t *testing.T) {
	t.Parallel()
	out, err := run(t, "--help")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Usage: echo") {
		t.Errorf("help out = %q", out)
	}
}

// TestVersionAsFirstArg verifies that a leading --version prints the version
// line rather than echoing the literal text.
func TestVersionAsFirstArg(t *testing.T) {
	t.Parallel()
	out, err := run(t, "--version")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "echo (mimixbox)") {
		t.Errorf("version out = %q", out)
	}
}

// TestHelpSections asserts `echo --help` renders structured help.
func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := echo.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Usage: echo", "Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q: %q", want, out.String())
		}
	}
}
