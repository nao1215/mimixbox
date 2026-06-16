package groups_test

import (
	"bytes"
	"context"
	"os/user"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/groups"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := groups.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// primaryGroupName returns the current user's primary group name, or "" if it
// cannot be resolved in this environment.
func primaryGroupName(t *testing.T) string {
	t.Helper()
	u, err := user.Current()
	if err != nil {
		t.Skipf("cannot determine current user: %v", err)
	}
	g, err := user.LookupGroupId(u.Gid)
	if err != nil {
		t.Skipf("cannot look up primary group %q: %v", u.Gid, err)
	}
	return g.Name
}

func TestRunNoArg(t *testing.T) {
	t.Parallel()
	want := primaryGroupName(t)

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatalf("expected non-empty output, got %q", out)
	}
	fields := strings.Fields(out)
	found := false
	for _, f := range fields {
		if f == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("output %q does not contain primary group %q", out, want)
	}
}

func TestRunUnknownUser(t *testing.T) {
	t.Parallel()
	const unknown = "no_such_user_zzqx_98765"

	out, errOut, err := run(t, unknown)
	if err == nil {
		t.Fatal("expected error for unknown user")
	}
	if !strings.Contains(errOut, "no such user") {
		t.Errorf("stderr = %q, want it to mention 'no such user'", errOut)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected no stdout for unknown user, got %q", out)
	}

	// The framework must turn this into a non-zero exit code.
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if code := command.Execute(context.Background(), groups.New(), io, []string{unknown}); code == command.ExitSuccess {
		t.Errorf("exit code = %d, want non-zero", code)
	}
}

func TestHelpSections(t *testing.T) {
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Examples:") || !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out)
	}
}
