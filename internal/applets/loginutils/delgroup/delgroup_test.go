package delgroup

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "group")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := groupPath
	groupPath = p
	t.Cleanup(func() { groupPath = orig })
	return p
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestRemovesGroup(t *testing.T) {
	p := fixture(t, "root:x:0:\nstaff:x:50:\ndevelopers:x:1000:\n")
	if err := run(t, "staff"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(p)
	if strings.Contains(string(data), "staff") {
		t.Errorf("staff should be removed:\n%s", data)
	}
	// The other groups must remain.
	if !strings.Contains(string(data), "root:x:0:") || !strings.Contains(string(data), "developers:x:1000:") {
		t.Errorf("other groups must be preserved:\n%s", data)
	}
}

func TestRemoveMatchesWholeName(t *testing.T) {
	// "staff" must not remove "staffroom".
	p := fixture(t, "staffroom:x:60:\n")
	if err := run(t, "staff"); err == nil {
		t.Errorf("a non-existent group should fail")
	}
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), "staffroom") {
		t.Errorf("staffroom must not be removed")
	}
}

func TestErrors(t *testing.T) {
	fixture(t, "root:x:0:\n")
	if err := run(t, "ghost"); err == nil {
		t.Errorf("removing a missing group should fail")
	}
	if err := run(t); err == nil {
		t.Errorf("missing group name should fail")
	}
}
