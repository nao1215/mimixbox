package applets

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// TestEveryAppletHelpHasExamplesAndExitStatus locks in the structured-help
// rollout (GitHub issues #491, #493): every registered applet must answer
// --help with a worked Examples section and an Exit status section, exiting 0.
// This is the regression guard that keeps newly added applets from shipping a
// bare option table.
func TestEveryAppletHelpHasExamplesAndExitStatus(t *testing.T) {
	t.Parallel()
	for name, applet := range Applets {
		name, applet := name, applet
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			out := &bytes.Buffer{}
			io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
			code := command.Execute(context.Background(), applet.Cmd, io, []string{"--help"})
			if code != command.ExitSuccess {
				t.Fatalf("%s --help exit = %d, want 0", name, code)
			}
			help := out.String()
			for _, want := range []string{"Examples:", "Exit status:"} {
				if !strings.Contains(help, want) {
					t.Errorf("%s --help is missing the %q section\n%s", name, want, help)
				}
			}
		})
	}
}
