package echo_test

import (
	"bytes"
	"context"
	"fmt"
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
		// GNU echo also takes \NNN without the leading zero.
		{"octal without leading zero", `\101`, "A\n"},
		{"octal one digit", `\1x`, "\x01x\n"},
		{"octal stops after three digits", `\1011`, "A1\n"},
		{"eight is not octal", `\8`, "\\8\n"},
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

// encodeEscapes writes every byte of data as an echo -e escape (or a literal),
// choosing the form per byte from styles. Each form is self-delimiting (\0NNN,
// \NNN and \xHH always use their full width), so a following literal digit is
// never absorbed and decoding must give data back.
func encodeEscapes(data, styles []byte) string {
	named := map[byte]string{'\a': `\a`, '\b': `\b`, '\f': `\f`, '\n': `\n`, '\r': `\r`, '\t': `\t`, '\v': `\v`, '\\': `\\`}
	var b strings.Builder
	for i, d := range data {
		style := byte(0)
		if len(styles) > 0 {
			style = styles[i%len(styles)] % 5
		}
		if i == 0 && d == '-' {
			style = 1 // a leading "-n" would be read as a flag, not text
		}
		switch {
		case style == 0 && d != '\\':
			b.WriteByte(d)
		case style == 2 && d >= 0o100:
			fmt.Fprintf(&b, `\%03o`, d)
		case style == 3:
			fmt.Fprintf(&b, `\x%02x`, d)
		case style == 4 && named[d] != "":
			b.WriteString(named[d])
		default:
			fmt.Fprintf(&b, `\0%03o`, d)
		}
	}
	return b.String()
}

// FuzzEscapeExpansion checks echo -e on arbitrary text and on encoded bytes.
// Arbitrary text must never panic, never grow (every escape is at least as long
// as the byte it stands for) and pass through unchanged when it has no
// backslash. Bytes encoded with the octal, hex and named escapes GNU echo
// accepts must decode back to themselves.
func FuzzEscapeExpansion(f *testing.F) {
	f.Add(`a\tb\tc`, []byte("A\x00\\-\n\xff7"), []byte{0, 1, 2, 3, 4})
	f.Add(`\0101\101\1x\8\x9z\xz\c`, []byte("-n"), []byte{0})
	f.Add(`end\`, []byte{}, []byte{})
	f.Fuzz(func(t *testing.T, raw string, data, styles []byte) {
		in := "x" + raw // keep the operand from looking like -n/-e/-E
		out, err := run(t, "-e", in)
		if err != nil {
			t.Fatalf("Run(-e, %q) error = %v", in, err)
		}
		if len(strings.TrimSuffix(out, "\n")) > len(in) {
			t.Fatalf("Run(-e, %q) = %q grew the input", in, out)
		}
		if !strings.Contains(in, `\`) && out != in+"\n" {
			t.Fatalf("Run(-e, %q) = %q, want the input unchanged", in, out)
		}

		enc := encodeEscapes(data, styles)
		out, err = run(t, "-n", "-e", enc)
		if err != nil {
			t.Fatalf("Run(-n, -e, %q) error = %v", enc, err)
		}
		if out != string(data) {
			t.Fatalf("Run(-n, -e, %q) = %q, want %q", enc, out, data)
		}
	})
}
