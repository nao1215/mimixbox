package paste_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/paste"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := paste.New().Run(context.Background(), io, args)
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

func TestSerial(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\nb\nc\n", "-s")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "a\tb\tc\n" {
		t.Errorf("out = %q", out)
	}
}

func TestSerialCustomDelimiter(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\nb\nc\n", "-s", "-d", ",")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "a,b,c\n" {
		t.Errorf("out = %q", out)
	}
}

func TestParallelTwoFiles(t *testing.T) {
	t.Parallel()
	f1 := writeFile(t, "1\n2\n3\n")
	f2 := writeFile(t, "a\nb\nc\n")
	out, _, err := run(t, "", f1, f2)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "1\ta\n2\tb\n3\tc\n" {
		t.Errorf("out = %q", out)
	}
}

func TestParallelUnevenLength(t *testing.T) {
	t.Parallel()
	f1 := writeFile(t, "1\n2\n")
	f2 := writeFile(t, "a\n")
	out, _, err := run(t, "", f1, f2)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "1\ta\n2\t\n" {
		t.Errorf("out = %q", out)
	}
}

// TestParallelRepeatedStdin covers the whole family of "- appears more than
// once" cases against GNU paste's behavior. Every `-` names the same stream, so
// consecutive stdin lines fill consecutive columns of one row; reading each
// operand to EOF in turn instead gives the first `-` everything and leaves the
// rest blank.
//
// The uneven cases are the boundary: the last row is short, and paste still
// emits it with the exhausted columns empty rather than dropping it.
func TestParallelRepeatedStdin(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		stdin string
		args  []string
		want  string
	}{
		"two columns consume two lines per row": {
			stdin: "a\nb\n",
			args:  []string{"-", "-"},
			want:  "a\tb\n",
		},
		"an odd line leaves the last column empty": {
			stdin: "a\nb\nc\n",
			args:  []string{"-", "-"},
			want:  "a\tb\nc\t\n",
		},
		"three columns consume three lines per row": {
			stdin: "a\nb\nc\nd\ne\n",
			args:  []string{"-", "-", "-"},
			want:  "a\tb\tc\nd\te\t\n",
		},
		"a custom delimiter still cycles across the columns": {
			stdin: "a\nb\nc\nd\n",
			args:  []string{"-d", ",", "-", "-"},
			want:  "a,b\nc,d\n",
		},
		"empty stdin produces no rows": {
			stdin: "",
			args:  []string{"-", "-"},
			want:  "",
		},
		"a single line fills only the first column": {
			stdin: "a\n",
			args:  []string{"-", "-"},
			want:  "a\t\n",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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

// TestParallelFileAndRepeatedStdin pins the mixed form: a real file supplies one
// column while the `-` operands share stdin, so the file's line count and the
// stdin line count advance independently.
func TestParallelFileAndRepeatedStdin(t *testing.T) {
	t.Parallel()

	file := writeFile(t, "x\ny\n")
	out, _, err := run(t, "a\nb\n", file, "-", "-")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "x\ta\tb\ny\t\t\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestSerialRepeatedStdin pins that -s is unaffected: it joins each operand's
// whole stream onto one line, so the first `-` consumes stdin and the second
// contributes an empty line. That already matched GNU and must stay that way.
func TestSerialRepeatedStdin(t *testing.T) {
	t.Parallel()

	out, _, err := run(t, "a\nb\n", "-s", "-", "-")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if want := "a\tb\n\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestDelimiterEscape(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\nb\n", "-s", "-d", `\n`)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "a\nb\n" {
		t.Errorf("out = %q", out)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "paste: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := paste.New()
	if c.Name() != "paste" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := paste.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Examples:") {
		t.Errorf("--help missing Examples section:\n%s", got)
	}
	if !strings.Contains(got, "Exit status:") {
		t.Errorf("--help missing Exit status section:\n%s", got)
	}
}
