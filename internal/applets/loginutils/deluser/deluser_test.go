package deluser

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixtures(t *testing.T, passwd, shadow, group string) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	op, os1, og := passwdPath, shadowPath, groupPath
	passwdPath = write("passwd", passwd)
	shadowPath = write("shadow", shadow)
	groupPath = write("group", group)
	t.Cleanup(func() { passwdPath, shadowPath, groupPath = op, os1, og })
	return passwdPath, shadowPath, groupPath
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, _ := os.ReadFile(path)
	return string(data)
}

func TestRemovesUserEverywhere(t *testing.T) {
	pw, sh, gr := fixtures(t,
		"root:x:0:0:root:/root:/bin/sh\nalice:x:1000:1000:alice:/home/alice:/bin/sh\n",
		"root:x:19000:0:99999:7:::\nalice:!:19000:0:99999:7:::\n",
		"root:x:0:\nstaff:x:50:alice,bob\nalice:x:1000:\n")
	if err := run(t, "alice"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(read(t, pw), "alice") {
		t.Errorf("alice should be gone from passwd:\n%s", read(t, pw))
	}
	if strings.Contains(read(t, sh), "alice:!") {
		t.Errorf("alice should be gone from shadow")
	}
	// alice removed from staff's member list, bob kept; the alice group line stays.
	if !strings.Contains(read(t, gr), "staff:x:50:bob") {
		t.Errorf("staff member list = %q", read(t, gr))
	}
	if !strings.Contains(read(t, pw), "root:x:0:") {
		t.Errorf("root must be preserved")
	}
}

func TestMembershipExactMatch(t *testing.T) {
	_, _, gr := fixtures(t,
		"al:x:1000:1000:al:/home/al:/bin/sh\n", "al:!:19000:0:99999:7:::\n",
		"team:x:60:al,alice,albert\n")
	if err := run(t, "al"); err != nil {
		t.Fatal(err)
	}
	// Only "al" removed; "alice" and "albert" kept.
	if got := read(t, gr); !strings.Contains(got, "team:x:60:alice,albert") {
		t.Errorf("member list = %q", got)
	}
}

func TestErrors(t *testing.T) {
	fixtures(t, "root:x:0:0:root:/root:/bin/sh\n", "root:x:19000:0:99999:7:::\n", "root:x:0:\n")
	if err := run(t, "ghost"); err == nil {
		t.Errorf("removing a missing user should fail")
	}
	if err := run(t); err == nil {
		t.Errorf("missing user name should fail")
	}
}
