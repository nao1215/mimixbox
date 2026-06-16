package groups_test

import (
	"os/user"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/groups"
)

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := groups.New()
	if c.Name() != "groups" {
		t.Errorf("Name() = %q, want %q", c.Name(), "groups")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestRunSingleUser covers the single-operand branch, which prints the group
// line without the "user :" prefix.
func TestRunSingleUser(t *testing.T) {
	t.Parallel()
	u, err := user.Current()
	if err != nil {
		t.Skipf("cannot determine current user: %v", err)
	}
	out, _, runErr := run(t, u.Username)
	if runErr != nil {
		t.Fatalf("Run error = %v", runErr)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatalf("expected non-empty output for user %q", u.Username)
	}
	// With a single operand there is no "user :" prefix.
	if strings.Contains(out, u.Username+" :") {
		t.Errorf("single-user output %q should not carry the 'user :' prefix", out)
	}
}

// TestRunMultipleUsers covers the multi-operand formatting branch, where each
// line is prefixed with "user :".
func TestRunMultipleUsers(t *testing.T) {
	t.Parallel()
	u, err := user.Current()
	if err != nil {
		t.Skipf("cannot determine current user: %v", err)
	}
	// The same user twice is a deterministic way to pass two valid operands.
	out, _, runErr := run(t, u.Username, u.Username)
	if runErr != nil {
		t.Fatalf("Run error = %v", runErr)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 output lines, got %d: %q", len(lines), out)
	}
	for _, line := range lines {
		if !strings.HasPrefix(line, u.Username+" : ") {
			t.Errorf("line %q does not start with %q prefix", line, u.Username+" : ")
		}
	}
}

// TestRunMixedValidAndInvalid covers the per-operand failure path: a valid user
// is still printed while an unknown user is reported and makes the run fail.
func TestRunMixedValidAndInvalid(t *testing.T) {
	t.Parallel()
	u, err := user.Current()
	if err != nil {
		t.Skipf("cannot determine current user: %v", err)
	}
	const unknown = "no_such_user_qqzz_55512"
	out, errOut, runErr := run(t, u.Username, unknown)
	if runErr == nil {
		t.Fatal("expected error when one operand is an unknown user")
	}
	if !strings.Contains(out, u.Username+" : ") {
		t.Errorf("output %q should still include the valid user's groups", out)
	}
	if !strings.Contains(errOut, "no such user") {
		t.Errorf("stderr = %q, want it to mention the unknown user", errOut)
	}
}
