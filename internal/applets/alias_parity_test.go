package applets

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// This file locks in the alias-parity contract (GitHub issue #789): the thin
// wrapper applets must not drift from the commands they delegate to.
//
//   - egrep is grep -E and fgrep is grep -F (internal/applets/findutils/grepalias
//     delegates to grep), so running the alias must match running grep with the
//     forced mode flag over the same stdin - identical stdout and identical exit
//     code, including on a shared error case.
//   - netcat is an alias of nc (internal/applets/netutils/netcat delegates to
//     nc); their --help must agree on the structured-help contract (same
//     sections, both exit 0, clean stderr) while each wrapper intentionally
//     renders its own command name, so the comparison is of the help body's
//     shape and semantics rather than its literal name lines.
//
// Tests reach the applets through the same registry and command.Execute path
// production uses, so they exercise the real dispatch rather than a private
// constructor.

// runApplet dispatches the registered applet under name with the given stdin and
// arguments, returning its stdout, stderr, and exit code through the production
// command.Execute path.
func runApplet(t *testing.T, name, stdin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	applet, ok := Applets[name]
	if !ok {
		t.Fatalf("applet %q is not registered", name)
	}
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	code = command.Execute(context.Background(), applet.Cmd, io, args)
	return out.String(), errBuf.String(), code
}

// TestEgrepMatchesGrepDashE asserts egrep PATTERN and grep -E PATTERN produce
// identical stdout and exit code over the same stdin.
func TestEgrepMatchesGrepDashE(t *testing.T) {
	t.Parallel()
	const in = "foo\nbar\nbaz\nqux\n"
	pattern := "ba(r|z)"

	aliasOut, _, aliasCode := runApplet(t, "egrep", in, pattern)
	baseOut, _, baseCode := runApplet(t, "grep", in, "-E", pattern)

	if aliasOut != baseOut {
		t.Errorf("egrep stdout = %q, grep -E stdout = %q; want identical", aliasOut, baseOut)
	}
	if aliasCode != baseCode {
		t.Errorf("egrep exit = %d, grep -E exit = %d; want identical", aliasCode, baseCode)
	}
}

// TestFgrepMatchesGrepDashF asserts fgrep PATTERN and grep -F PATTERN produce
// identical stdout and exit code over the same stdin. The pattern contains a
// regex metacharacter so the fixed-string mode is observably different from
// extended mode.
func TestFgrepMatchesGrepDashF(t *testing.T) {
	t.Parallel()
	const in = "a.b\naxb\na.b.c\n"
	pattern := "a.b"

	aliasOut, _, aliasCode := runApplet(t, "fgrep", in, pattern)
	baseOut, _, baseCode := runApplet(t, "grep", in, "-F", pattern)

	if aliasOut != baseOut {
		t.Errorf("fgrep stdout = %q, grep -F stdout = %q; want identical", aliasOut, baseOut)
	}
	if aliasCode != baseCode {
		t.Errorf("fgrep exit = %d, grep -F exit = %d; want identical", aliasCode, baseCode)
	}
}

// TestEgrepErrorExitParity asserts egrep and grep -E agree on a non-success exit
// code for a shared error case: a pattern against a file that does not exist.
func TestEgrepErrorExitParity(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "does-not-exist.txt")

	_, _, aliasCode := runApplet(t, "egrep", "", "foo", missing)
	_, _, baseCode := runApplet(t, "grep", "", "-E", "foo", missing)

	if aliasCode == command.ExitSuccess {
		t.Errorf("egrep on a missing file exited 0, want non-zero")
	}
	if aliasCode != baseCode {
		t.Errorf("egrep exit = %d, grep -E exit = %d; want identical on the error case", aliasCode, baseCode)
	}
}

// TestFgrepErrorExitParity is the fixed-string counterpart of
// TestEgrepErrorExitParity.
func TestFgrepErrorExitParity(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "does-not-exist.txt")

	_, _, aliasCode := runApplet(t, "fgrep", "", "foo", missing)
	_, _, baseCode := runApplet(t, "grep", "", "-F", "foo", missing)

	if aliasCode == command.ExitSuccess {
		t.Errorf("fgrep on a missing file exited 0, want non-zero")
	}
	if aliasCode != baseCode {
		t.Errorf("fgrep exit = %d, grep -F exit = %d; want identical on the error case", aliasCode, baseCode)
	}
}

// TestNetcatHelpParityWithNc asserts that netcat --help and nc --help honor the
// same structured-help contract: both exit 0, write nothing to stderr, and carry
// the same sections (Usage, Examples, Exit status). The one intentional
// difference is the command name each wrapper prints, so this test compares the
// help body's shape - not the literal name line - and confirms each wrapper
// renders its own name in the Usage line (and does not leak the other's). No
// sockets are opened: --help short-circuits before any network code runs.
func TestNetcatHelpParityWithNc(t *testing.T) {
	t.Parallel()
	netcatOut, netcatErr, netcatCode := runApplet(t, "netcat", "", "--help")
	ncOut, ncErr, ncCode := runApplet(t, "nc", "", "--help")

	if netcatCode != command.ExitSuccess {
		t.Errorf("netcat --help exit = %d, want 0", netcatCode)
	}
	if ncCode != command.ExitSuccess {
		t.Errorf("nc --help exit = %d, want 0", ncCode)
	}
	if netcatErr != "" {
		t.Errorf("netcat --help wrote to stderr: %q", netcatErr)
	}
	if ncErr != "" {
		t.Errorf("nc --help wrote to stderr: %q", ncErr)
	}

	// Each wrapper renders its OWN name in the Usage line and never the other's.
	if !strings.HasPrefix(netcatOut, "Usage: netcat") {
		t.Errorf("netcat --help should start with %q:\n%s", "Usage: netcat", netcatOut)
	}
	if !strings.HasPrefix(ncOut, "Usage: nc") {
		t.Errorf("nc --help should start with %q:\n%s", "Usage: nc", ncOut)
	}
	if strings.Contains(ncOut, "netcat") {
		t.Errorf("nc --help leaks the alias name %q:\n%s", "netcat", ncOut)
	}

	// Structured-help parity: the same sections appear in both, so neither
	// wrapper drifts to a bare option dump. This is the body/semantics comparison
	// the contract calls for, independent of the per-command name and prose.
	for _, section := range []string{"Usage:", "Options:", "Examples:", "Exit status:"} {
		if !strings.Contains(netcatOut, section) {
			t.Errorf("netcat --help missing %q section:\n%s", section, netcatOut)
		}
		if !strings.Contains(ncOut, section) {
			t.Errorf("nc --help missing %q section:\n%s", section, ncOut)
		}
	}
}
