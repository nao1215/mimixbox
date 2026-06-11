package chpasswd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GehirnInc/crypt"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixtureShadow(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "shadow")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	orig := shadowPath
	shadowPath = p
	t.Cleanup(func() { shadowPath = orig })
	return p
}

func run(t *testing.T, stdin string, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(stdin), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func field(t *testing.T, path, user string, idx int) string {
	t.Helper()
	data, _ := os.ReadFile(path)
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		f := strings.Split(line, ":")
		if f[0] == user {
			return f[idx]
		}
	}
	t.Fatalf("user %s not found", user)
	return ""
}

func TestHashesAndUpdates(t *testing.T) {
	p := fixtureShadow(t, "alice:!:19000:0:99999:7:::\nbob:!:19000:0:99999:7:::\n")
	if err := run(t, "alice:newsecret\n"); err != nil {
		t.Fatal(err)
	}
	hash := field(t, p, "alice", 1)
	if !strings.HasPrefix(hash, "$6$") {
		t.Errorf("alice hash = %q, want $6$ sha-512", hash)
	}
	if crypt.NewFromHash(hash).Verify(hash, []byte("newsecret")) != nil {
		t.Errorf("alice's new hash does not verify against the password")
	}
	// bob is untouched, and the other fields are preserved.
	if field(t, p, "bob", 1) != "!" {
		t.Errorf("bob should be unchanged")
	}
	if field(t, p, "alice", 4) != "99999" {
		t.Errorf("alice's other shadow fields must be preserved")
	}
}

func TestEncryptedStoredVerbatim(t *testing.T) {
	p := fixtureShadow(t, "alice:!:19000:0:99999:7:::\n")
	if err := run(t, "alice:$6$prehashed$abc\n", "-e"); err != nil {
		t.Fatal(err)
	}
	if got := field(t, p, "alice", 1); got != "$6$prehashed$abc" {
		t.Errorf("encrypted value = %q, want stored verbatim", got)
	}
}

func TestMD5Method(t *testing.T) {
	p := fixtureShadow(t, "alice:!:19000:0:99999:7:::\n")
	if err := run(t, "alice:secret\n", "-c", "md5"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(field(t, p, "alice", 1), "$1$") {
		t.Errorf("md5 method should produce a $1$ hash")
	}
}

func TestUnknownUser(t *testing.T) {
	fixtureShadow(t, "alice:!:19000:0:99999:7:::\n")
	if err := run(t, "carol:secret\n"); err == nil {
		t.Errorf("an unknown user should fail")
	}
}

func TestErrors(t *testing.T) {
	fixtureShadow(t, "alice:!::::::: \n")
	if err := run(t, "alice:secret\n", "-c", "bogus"); err == nil {
		t.Errorf("unknown method should fail")
	}
	if err := run(t, "no-colon-line\n"); err == nil {
		t.Errorf("a malformed input line should fail")
	}
}
