package adduser

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

func lineFor(t *testing.T, path, name string) string {
	t.Helper()
	data, _ := os.ReadFile(path)
	for _, l := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if strings.HasPrefix(l, name+":") {
			return l
		}
	}
	return ""
}

func TestCreatesUserAndGroup(t *testing.T) {
	pw, sh, gr := fixtures(t, "root:x:0:0:root:/root:/bin/sh\n", "root:x:19000:0:99999:7:::\n", "root:x:0:\n")
	if err := run(t, "alice"); err != nil {
		t.Fatal(err)
	}
	// passwd: alice with auto UID 1000, GID 1000, default home/shell.
	if got := lineFor(t, pw, "alice"); got != "alice:x:1000:1000:alice:/home/alice:/bin/sh" {
		t.Errorf("passwd line = %q", got)
	}
	// shadow: locked account.
	if got := lineFor(t, sh, "alice"); !strings.HasPrefix(got, "alice:!:") {
		t.Errorf("shadow line = %q, want locked", got)
	}
	// a matching group was created.
	if got := lineFor(t, gr, "alice"); got != "alice:x:1000:" {
		t.Errorf("group line = %q", got)
	}
}

func TestUsesExistingGroup(t *testing.T) {
	pw, _, _ := fixtures(t, "root:x:0:0:root:/root:/bin/sh\n", "root:x:19000:0:99999:7:::\n", "root:x:0:\nstaff:x:50:\n")
	if err := run(t, "-G", "staff", "-s", "/bin/bash", "bob"); err != nil {
		t.Fatal(err)
	}
	got := lineFor(t, pw, "bob")
	if !strings.Contains(got, ":1000:50:") || !strings.HasSuffix(got, ":/bin/bash") {
		t.Errorf("passwd line = %q, want GID 50 and bash", got)
	}
}

func TestSpecificUID(t *testing.T) {
	pw, _, _ := fixtures(t, "root:x:0:0:root:/root:/bin/sh\n", "root:x:19000:0:99999:7:::\n", "root:x:0:\n")
	if err := run(t, "-u", "2000", "-h", "/srv/carol", "carol"); err != nil {
		t.Fatal(err)
	}
	got := lineFor(t, pw, "carol")
	if !strings.HasPrefix(got, "carol:x:2000:") || !strings.Contains(got, ":/srv/carol:") {
		t.Errorf("passwd line = %q", got)
	}
}

func TestErrors(t *testing.T) {
	fixtures(t, "alice:x:1000:1000:alice:/home/alice:/bin/sh\n", "alice:!:19000:0:99999:7:::\n", "alice:x:1000:\n")
	if err := run(t, "alice"); err == nil {
		t.Errorf("an existing user should fail")
	}
	if err := run(t, "-u", "1000", "newuser"); err == nil {
		t.Errorf("an in-use UID should fail")
	}
	if err := run(t, "-G", "nosuchgroup", "newuser"); err == nil {
		t.Errorf("a missing primary group should fail")
	}
	if err := run(t); err == nil {
		t.Errorf("missing user name should fail")
	}
}
