package applets

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// requiredHelpSections is the opted-in help-shape policy (GitHub issue #786):
// every applet's --help must carry both a worked Examples block and an Exit
// status block, not just a bare option table.
var requiredHelpSections = []string{"Examples:", "Exit status:"}

// missingHelpSections returns the policy sections that help does NOT contain.
// An empty result means help satisfies the policy. Factoring the check here lets
// the real applets and the negative fixture be judged by the exact same rule.
func missingHelpSections(help string) []string {
	var missing []string
	for _, section := range requiredHelpSections {
		if !strings.Contains(help, section) {
			missing = append(missing, section)
		}
	}
	return missing
}

// TestEveryAppletHelpSatisfiesShapePolicy asserts that every registered applet's
// --help passes the help-shape policy.
func TestEveryAppletHelpSatisfiesShapePolicy(t *testing.T) {
	t.Parallel()
	for name := range Applets {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			help, _, code := runApplet(t, name, "", "--help")
			if code != command.ExitSuccess {
				t.Fatalf("%s --help exit = %d, want %d", name, code, command.ExitSuccess)
			}
			if missing := missingHelpSections(help); len(missing) > 0 {
				t.Errorf("%s --help violates help-shape policy: missing %v\n%s", name, missing, help)
			}
		})
	}
}

// bareUsageCmd is a negative fixture: a command whose --help renders only a bare
// "Usage:"/options block, with neither an Examples nor an Exit status section.
// It exists to prove that missingHelpSections actually rejects an unadorned
// usage block, so the policy test above cannot silently pass on everything.
type bareUsageCmd struct{}

func (bareUsageCmd) Name() string     { return "fixturecmd" }
func (bareUsageCmd) Synopsis() string { return "negative fixture command" }

func (bareUsageCmd) Run(_ context.Context, io command.IO, _ []string) error {
	// Render the GNU-style usage block without attaching any WithHelp sections,
	// so the output is exactly the bare "Usage:"/Options shape the policy must
	// reject.
	errBuf := &bytes.Buffer{}
	fs := command.NewFlagSet("fixturecmd", "", errBuf)
	fs.WriteUsage(io.Out)
	return nil
}

// TestHelpShapePolicyRejectsBareUsageFixture verifies the policy detects a bare
// usage block: the negative fixture's --help must be reported as missing BOTH
// required sections.
func TestHelpShapePolicyRejectsBareUsageFixture(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	code := command.Execute(context.Background(), bareUsageCmd{}, io, []string{"--help"})
	if code != command.ExitSuccess {
		t.Fatalf("fixture --help exit = %d, want %d", code, command.ExitSuccess)
	}

	help := out.String()
	if !strings.Contains(help, "Usage: fixturecmd") {
		t.Fatalf("fixture --help should still print a usage line, got:\n%s", help)
	}

	missing := missingHelpSections(help)
	if len(missing) != len(requiredHelpSections) {
		t.Fatalf("policy should reject the bare usage fixture for all %d sections, but only flagged %v\n%s",
			len(requiredHelpSections), missing, help)
	}
}
