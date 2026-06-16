package unshadow

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestMerge(t *testing.T) {
	t.Parallel()
	passwd := writeFile(t, "passwd",
		"root:x:0:0:root:/root:/bin/bash\nalice:x:1000:1000:Alice:/home/alice:/bin/sh\n")
	shadow := writeFile(t, "shadow",
		"root:$6$abc$HASHROOT:19000:0:99999:7:::\nalice:$6$def$HASHALICE:19000:0:99999:7:::\n")
	out, _, err := run(t, passwd, shadow)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "root:$6$abc$HASHROOT:0:0:root:/root:/bin/bash") {
		t.Errorf("root line not merged: %q", out)
	}
	if !strings.Contains(out, "alice:$6$def$HASHALICE:1000:1000:Alice:/home/alice:/bin/sh") {
		t.Errorf("alice line not merged: %q", out)
	}
}

func TestUserMissingFromShadowKeepsPasswdField(t *testing.T) {
	t.Parallel()
	passwd := writeFile(t, "passwd", "bob:x:1001:1001::/home/bob:/bin/sh\n")
	shadow := writeFile(t, "shadow", "root:$6$x$H:19000:0:99999:7:::\n")
	out, _, err := run(t, passwd, shadow)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "bob:x:1001:1001") {
		t.Errorf("bob's field should be unchanged: %q", out)
	}
}

func TestWrongOperandCount(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "only-one")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "two file operands") {
		t.Errorf("err = %v", err)
	}
}

func TestMissingPasswdFile(t *testing.T) {
	t.Parallel()
	shadow := writeFile(t, "shadow", "x:y:1::\n")
	_, _, err := run(t, "/no/such/passwd", shadow)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMissingShadowFile(t *testing.T) {
	t.Parallel()
	passwd := writeFile(t, "passwd", "root:x:0:0::/root:/bin/sh\n")
	_, _, err := run(t, passwd, "/no/such/shadow")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "unshadow" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	if !strings.Contains(out.String(), "Examples:") || !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out.String())
	}
}
