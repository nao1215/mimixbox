package applets

import (
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// TestEveryAppletSatisfiesHelpVersionContract is the registry-wide contract
// (GitHub issue #785): every registered applet must answer --help and --version
// the same way, so a user (or a tool) can rely on the conventions regardless of
// which applet they invoke.
//
//   - --help exits 0 and prints a "Usage: <name>" line.
//   - --version exits 0 and prints the "<name> (mimixbox) <version>" banner.
//
// This complements TestEveryAppletHelpHasExamplesAndExitStatus, which checks the
// richer help sections; here we lock in the bare usage/version surface itself.
func TestEveryAppletSatisfiesHelpVersionContract(t *testing.T) {
	t.Parallel()
	for name := range Applets {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			help, _, code := runApplet(t, name, "", "--help")
			if code != command.ExitSuccess {
				t.Fatalf("%s --help exit = %d, want %d", name, code, command.ExitSuccess)
			}
			if want := "Usage: " + name; !strings.Contains(help, want) {
				t.Errorf("%s --help is missing usage text %q\n%s", name, want, help)
			}

			version, _, code := runApplet(t, name, "", "--version")
			if code != command.ExitSuccess {
				t.Fatalf("%s --version exit = %d, want %d", name, code, command.ExitSuccess)
			}
			if want := " (mimixbox) "; !strings.Contains(version, want) {
				t.Errorf("%s --version is missing the %q banner\n%s", name, want, version)
			}
		})
	}
}
