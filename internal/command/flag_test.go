package command_test

import (
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestWriteUsageRendersAllHelpSections(t *testing.T) {
	t.Parallel()
	io, out, _ := newIO()
	fs := command.NewFlagSet("demo", "[OPTION]... [FILE]...", io.Err).WithHelp(command.Help{
		Description: "Demo concatenates files.",
		Examples: []command.Example{
			{Command: "demo a.txt", Explain: "print a.txt"},
			{Command: "demo -n a.txt b.txt", Explain: "number lines of both files"},
		},
		ExitStatus: "0  success\n1  failure",
		Notes:      []string{"first note", "second note"},
	})
	fs.BoolP("number", "n", false, "number all output lines")

	proceed, err := fs.Parse(io, []string{"--help"})
	if err != nil || proceed {
		t.Fatalf("Parse(--help) = (%v, %v), want (false, nil)", proceed, err)
	}

	got := out.String()
	wants := []string{
		"Usage: demo [OPTION]... [FILE]...",
		"Demo concatenates files.",
		"Options:",
		"--number",
		"Examples:",
		"demo a.txt",
		"print a.txt",
		"demo -n a.txt b.txt",
		"number lines of both files",
		"Exit status:",
		"  0  success",
		"  1  failure",
		"Notes:",
		"  - first note",
		"  - second note",
	}
	for _, w := range wants {
		if !strings.Contains(got, w) {
			t.Errorf("usage output missing %q\nfull output:\n%s", w, got)
		}
	}
}

func TestWriteUsageExampleColumnsAreAligned(t *testing.T) {
	t.Parallel()
	io, out, _ := newIO()
	fs := command.NewFlagSet("demo", "", io.Err).WithHelp(command.Help{
		Examples: []command.Example{
			{Command: "short", Explain: "A"},
			{Command: "a-much-longer-command", Explain: "B"},
		},
	})

	if _, err := fs.Parse(io, []string{"--help"}); err != nil {
		t.Fatalf("Parse(--help) error = %v", err)
	}

	// The short command must be padded to the width of the longest command so
	// the explanations line up in the same column on every line.
	colOf := func(line, explain string) int {
		i := strings.Index(line, explain)
		if i < 0 {
			t.Fatalf("explanation %q not found in line %q", explain, line)
		}
		return i
	}
	var shortLine, longLine string
	for _, line := range strings.Split(out.String(), "\n") {
		switch {
		case strings.Contains(line, "short") && strings.HasSuffix(line, "A"):
			shortLine = line
		case strings.Contains(line, "a-much-longer-command"):
			longLine = line
		}
	}
	if shortLine == "" || longLine == "" {
		t.Fatalf("example lines not found:\n%s", out.String())
	}
	if colOf(shortLine, "A") != colOf(longLine, "B") {
		t.Errorf("explanations are not column-aligned:\n%q\n%q", shortLine, longLine)
	}
}

func TestWriteUsageOmitsEmptySections(t *testing.T) {
	t.Parallel()
	io, out, _ := newIO()
	// No WithHelp: only Usage and Options should appear.
	fs := command.NewFlagSet("demo", "", io.Err)

	if _, err := fs.Parse(io, []string{"--help"}); err != nil {
		t.Fatalf("Parse(--help) error = %v", err)
	}

	got := out.String()
	for _, absent := range []string{"Examples:", "Exit status:", "Notes:"} {
		if strings.Contains(got, absent) {
			t.Errorf("usage output should not contain %q when no help is attached:\n%s", absent, got)
		}
	}
	if !strings.Contains(got, "Options:") {
		t.Errorf("usage output should always contain Options:\n%s", got)
	}
}

func TestParseEndOfOptionsAndInterspersed(t *testing.T) {
	t.Parallel()
	io, _, _ := newIO()
	fs := command.NewFlagSet("demo", "", io.Err)
	n := fs.BoolP("number", "n", false, "")

	// "--" ends option parsing: the "-x" after it is an operand, not a flag.
	proceed, err := fs.Parse(io, []string{"-n", "file.txt", "--", "-x"})
	if err != nil || !proceed {
		t.Fatalf("Parse = (%v, %v), want (true, nil)", proceed, err)
	}
	if !*n {
		t.Errorf("-n should be set")
	}
	args := fs.Args()
	if len(args) != 2 || args[0] != "file.txt" || args[1] != "-x" {
		t.Errorf("operands = %v, want [file.txt -x]", args)
	}
}

func TestParseProceedsWhenNoControlFlag(t *testing.T) {
	t.Parallel()
	io, out, errBuf := newIO()
	fs := command.NewFlagSet("demo", "", io.Err)

	proceed, err := fs.Parse(io, []string{"operand"})
	if err != nil || !proceed {
		t.Fatalf("Parse = (%v, %v), want (true, nil)", proceed, err)
	}
	if out.Len() != 0 || errBuf.Len() != 0 {
		t.Errorf("Parse without help/version should be silent; out=%q err=%q", out.String(), errBuf.String())
	}
}

func TestWithHelpReturnsSameFlagSet(t *testing.T) {
	t.Parallel()
	io, _, _ := newIO()
	fs := command.NewFlagSet("demo", "", io.Err)
	if got := fs.WithHelp(command.Help{Description: "x"}); got != fs {
		t.Errorf("WithHelp should return the same FlagSet for chaining")
	}
}
