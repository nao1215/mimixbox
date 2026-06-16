package env_test

import (
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/env"
)

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := env.New()
	if c.Name() != "env" {
		t.Errorf("Name() = %q, want %q", c.Name(), "env")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestRunAssignmentReplacesExisting drives setEnv's replace-an-existing-entry
// branch: a later NAME=VALUE for the same name must overwrite the earlier one,
// and only the final value may appear in the printed environment.
func TestRunAssignmentReplacesExisting(t *testing.T) {
	out, _, err := run(t, "", "-i", "DUP=first", "DUP=second")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "DUP=second") {
		t.Errorf("output %q does not contain final assignment DUP=second", out)
	}
	if strings.Contains(out, "DUP=first") {
		t.Errorf("output %q still contains the overwritten DUP=first", out)
	}
	if n := strings.Count(out, "DUP="); n != 1 {
		t.Errorf("DUP= appears %d times, want exactly 1: %q", n, out)
	}
}

// TestRunReplacesInheritedVariable confirms that a NAME=VALUE operand replaces a
// variable already present in the inherited environment rather than appending a
// duplicate.
func TestRunReplacesInheritedVariable(t *testing.T) {
	t.Setenv("MIMIX_OVERRIDE", "inherited")
	out, _, err := run(t, "", "MIMIX_OVERRIDE=overridden")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "MIMIX_OVERRIDE=overridden") {
		t.Errorf("output %q does not contain the override value", out)
	}
	if strings.Contains(out, "MIMIX_OVERRIDE=inherited") {
		t.Errorf("output %q still contains the inherited value", out)
	}
	if n := strings.Count(out, "MIMIX_OVERRIDE="); n != 1 {
		t.Errorf("MIMIX_OVERRIDE= appears %d times, want exactly 1", n)
	}
}

// TestRunNonAssignmentStartsCommand verifies that an operand without '=' is the
// start of the command, not an assignment: "=x" has an empty name so it is not
// an assignment and is treated as the (missing) command, which fails to start.
func TestRunBareEqualsIsCommand(t *testing.T) {
	_, errOut, err := run(t, "", "-i", "=novalue")
	if err == nil {
		t.Fatal("expected error: '=novalue' should be treated as a command, not an assignment")
	}
	if !strings.Contains(errOut, "env: '=novalue'") {
		t.Errorf("stderr = %q, want it to treat =novalue as a command", errOut)
	}
}
