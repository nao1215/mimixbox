package passwd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "shadow")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	os1, ou := shadowPath, currentUserFn
	shadowPath = p
	currentUserFn = func() (string, error) { return "tester", nil }
	t.Cleanup(func() { shadowPath, currentUserFn = os1, ou })
	return p
}

func run(t *testing.T, stdin string, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(stdin), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func field(t *testing.T, path, user string, idx int) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading shadow: %v", err)
	}
	for _, l := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		f := strings.Split(l, ":")
		if len(f) > idx && f[0] == user {
			return f[idx]
		}
	}
	t.Fatalf("user %s (field %d) not found", user, idx)
	return ""
}

func hashField(t *testing.T, path, user string) string {
	t.Helper()
	return field(t, path, user, 1)
}

func TestSetPassword(t *testing.T) {
	p := fixture(t, "alice:!:19000:0:99999:7:::\n")
	if err := run(t, "newsecret\n", "alice"); err != nil {
		t.Fatal(err)
	}
	h := hashField(t, p, "alice")
	if !strings.HasPrefix(h, "$6$") || crypt.NewFromHash(h).Verify(h, []byte("newsecret")) != nil {
		t.Errorf("hash %q does not verify", h)
	}
}

func TestPasswordMismatch(t *testing.T) {
	fixture(t, "alice:!:19000:0:99999:7:::\n")
	if err := run(t, "one\ntwo\n", "alice"); err == nil {
		t.Errorf("mismatched confirmation should fail")
	}
}

func TestLockUnlock(t *testing.T) {
	p := fixture(t, "alice:$6$abc$def:19000:0:99999:7:::\n")
	if err := run(t, "", "-l", "alice"); err != nil {
		t.Fatal(err)
	}
	if got := hashField(t, p, "alice"); got != "!$6$abc$def" {
		t.Errorf("locked hash = %q", got)
	}
	if err := run(t, "", "-u", "alice"); err != nil {
		t.Fatal(err)
	}
	if got := hashField(t, p, "alice"); got != "$6$abc$def" {
		t.Errorf("unlocked hash = %q", got)
	}
}

func TestDelete(t *testing.T) {
	p := fixture(t, "alice:$6$abc$def:19000:0:99999:7:::\n")
	if err := run(t, "", "-d", "alice"); err != nil {
		t.Fatal(err)
	}
	if got := hashField(t, p, "alice"); got != "" {
		t.Errorf("deleted password field = %q, want empty", got)
	}
}

func TestDefaultsToCurrentUser(t *testing.T) {
	p := fixture(t, "tester:!:19000:0:99999:7:::\n")
	if err := run(t, "mypw\n"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hashField(t, p, "tester"), "$6$") {
		t.Errorf("current user's password not set")
	}
}

func TestSetUpdatesLastChange(t *testing.T) {
	on := now
	now = func() time.Time { return time.Unix(20000*86400, 0) } // day 20000
	defer func() { now = on }()
	p := fixture(t, "alice:!:1:0:99999:7:::\n")
	if err := run(t, "newsecret\n", "alice"); err != nil {
		t.Fatal(err)
	}
	if got := field(t, p, "alice", 2); got != "20000" {
		t.Errorf("lastchg = %q, want 20000", got)
	}
}

func TestRejectsExtraArgs(t *testing.T) {
	fixture(t, "alice:!:19000:0:99999:7:::\n")
	if err := run(t, "x\n", "alice", "bob"); err == nil {
		t.Errorf("extra positional args should fail")
	}
}

func TestErrors(t *testing.T) {
	fixture(t, "alice:!:19000:0:99999:7:::\n")
	if err := run(t, "x\n", "ghost"); err == nil {
		t.Errorf("an unknown user should fail")
	}
	if err := run(t, "", "-l", "-u", "alice"); err == nil {
		t.Errorf("conflicting flags should fail")
	}
	if err := run(t, "\n", "alice"); err == nil {
		t.Errorf("an empty password should fail")
	}
}

func TestUsesStableLockfile(t *testing.T) {
	p := fixture(t, "alice:!:19000:0:99999:7:::\n")
	if err := run(t, "newsecret\n", "alice"); err != nil {
		t.Fatal(err)
	}
	// The lock is taken on a dedicated lockfile next to the shadow file, which
	// survives the atomic rename of the shadow file's inode.
	if _, err := os.Stat(p + ".lock"); err != nil {
		t.Errorf("expected a %s.lock lockfile: %v", p, err)
	}
}
