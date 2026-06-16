package id_test

import (
	"bytes"
	"context"
	"os"
	"os/user"
	"strconv"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/id"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := id.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunUserID(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "-u")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	got := strings.TrimSpace(out)
	want := strconv.Itoa(os.Getuid())
	if got != want {
		t.Errorf("uid = %q, want %q", got, want)
	}
}

func TestRunDefault(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "uid=") {
		t.Errorf("default output = %q, want to contain %q", out, "uid=")
	}
}

func TestRunUserName(t *testing.T) {
	t.Parallel()
	cur, err := user.Current()
	if err != nil {
		t.Skipf("cannot determine current user: %v", err)
	}
	out, errOut, runErr := run(t, "-u", "-n")
	if runErr != nil {
		t.Fatalf("Run error = %v (stderr=%q)", runErr, errOut)
	}
	got := strings.TrimSpace(out)
	if got != cur.Username {
		t.Errorf("user name = %q, want %q", got, cur.Username)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := id.New()
	if c.Name() != "id" {
		t.Errorf("Name() = %q, want %q", c.Name(), "id")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestRunGroupID exercises dumpGID: -g prints the primary group ID.
func TestRunGroupID(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "-g")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	got := strings.TrimSpace(out)
	want := strconv.Itoa(os.Getgid())
	if got != want {
		t.Errorf("gid = %q, want %q", got, want)
	}
}

// TestRunGroupName exercises dumpGID's showName branch with -g -n.
func TestRunGroupName(t *testing.T) {
	t.Parallel()
	cur, err := user.Current()
	if err != nil {
		t.Skipf("cannot determine current user: %v", err)
	}
	g, err := user.LookupGroupId(cur.Gid)
	if err != nil {
		t.Skipf("cannot look up primary group: %v", err)
	}
	out, errOut, runErr := run(t, "-g", "-n")
	if runErr != nil {
		t.Fatalf("Run error = %v (stderr=%q)", runErr, errOut)
	}
	if strings.TrimSpace(out) != g.Name {
		t.Errorf("group name = %q, want %q", strings.TrimSpace(out), g.Name)
	}
}

// TestRunAllGroups exercises dumpGroups: -G prints the supplementary group IDs,
// which must include the primary GID.
func TestRunAllGroups(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "-G")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, strconv.Itoa(os.Getgid())) {
		t.Errorf("groups = %q, want to contain primary gid %d", out, os.Getgid())
	}
}

// TestRunAllGroupNames exercises dumpGroups' showName branch with -G -n.
func TestRunAllGroupNames(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "-G", "-n")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("group names output is empty")
	}
}

// TestRunNamedUser exercises resolveUser's lookup path with an explicit operand.
func TestRunNamedUser(t *testing.T) {
	t.Parallel()
	cur, err := user.Current()
	if err != nil {
		t.Skipf("cannot determine current user: %v", err)
	}
	out, errOut, runErr := run(t, "-u", cur.Username)
	if runErr != nil {
		t.Fatalf("Run error = %v (stderr=%q)", runErr, errOut)
	}
	if strings.TrimSpace(out) != cur.Uid {
		t.Errorf("uid = %q, want %q", strings.TrimSpace(out), cur.Uid)
	}
}

func TestRunNoSuchUser(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "definitely-no-such-user-xyz")
	if err == nil {
		t.Fatal("expected error for unknown user")
	}
	if !strings.Contains(errOut, "no such user") {
		t.Errorf("stderr = %q, want no-such-user message", errOut)
	}
}

func TestRunExtraOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "alice", "bob")
	if err == nil {
		t.Fatal("expected error for extra operand")
	}
	if !strings.Contains(errOut, "extra operand") {
		t.Errorf("stderr = %q, want extra-operand message", errOut)
	}
}

func TestRunConflictingSelectors(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "-u", "-g")
	if err == nil {
		t.Fatal("expected error for more than one of -u/-g/-G")
	}
	if !strings.Contains(errOut, "more than one choice") {
		t.Errorf("stderr = %q, want conflicting-choice message", errOut)
	}
}

func TestRunNameWithoutSelector(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "-n")
	if err == nil {
		t.Fatal("expected error for -n without -u/-g/-G")
	}
	if !strings.Contains(errOut, "default format") {
		t.Errorf("stderr = %q, want default-format message", errOut)
	}
}
