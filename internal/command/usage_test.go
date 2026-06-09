package command_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// TestWriteUsageSections verifies that WithHelp's sections (description,
// examples, exit status, notes) are rendered in --help, and that a plain
// FlagSet with no rich help still renders just the Usage/Options block.
func TestWriteUsageSections(t *testing.T) {
	t.Parallel()

	t.Run("rich help renders every section", func(t *testing.T) {
		t.Parallel()
		var errBuf bytes.Buffer
		fs := command.NewFlagSet("demo", "[OPTION]... FILE", &errBuf).WithHelp(command.Help{
			Description: "Demo does a thing.",
			Examples: []command.Example{
				{Command: "demo a", Explain: "Do it to a."},
				{Command: "demo -x b", Explain: "Do it to b with -x."},
			},
			ExitStatus: "0  ok.\n1  failure.",
			Notes:      []string{"This is a note.", "Another note."},
		})

		var out bytes.Buffer
		fs.WriteUsage(&out)
		got := out.String()

		for _, want := range []string{
			"Usage: demo [OPTION]... FILE",
			"Demo does a thing.",
			"Options:",
			"--help",
			"Examples:",
			"demo a",
			"Do it to a.",
			"demo -x b",
			"Exit status:",
			"  0  ok.",
			"Notes:",
			"  - This is a note.",
			"  - Another note.",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("usage output is missing %q\n--- got ---\n%s", want, got)
			}
		}
	})

	t.Run("examples align on the widest command", func(t *testing.T) {
		t.Parallel()
		var errBuf bytes.Buffer
		fs := command.NewFlagSet("demo", "", &errBuf).WithHelp(command.Help{
			Examples: []command.Example{
				{Command: "a", Explain: "short"},
				{Command: "longcommand", Explain: "long"},
			},
		})
		var out bytes.Buffer
		fs.WriteUsage(&out)
		got := out.String()
		// The short command is padded to the width of the longest one.
		if !strings.Contains(got, "  a            short") {
			t.Errorf("examples not aligned:\n%s", got)
		}
	})

	t.Run("no rich help renders only usage and options", func(t *testing.T) {
		t.Parallel()
		var errBuf bytes.Buffer
		fs := command.NewFlagSet("bare", "[FILE]", &errBuf)
		var out bytes.Buffer
		fs.WriteUsage(&out)
		got := out.String()
		if !strings.Contains(got, "Usage: bare [FILE]") || !strings.Contains(got, "Options:") {
			t.Errorf("bare usage missing core sections:\n%s", got)
		}
		for _, unwanted := range []string{"Examples:", "Exit status:", "Notes:"} {
			if strings.Contains(got, unwanted) {
				t.Errorf("bare usage should not contain %q:\n%s", unwanted, got)
			}
		}
	})
}
