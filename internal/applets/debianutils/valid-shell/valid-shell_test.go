package validShell_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	validShell "github.com/nao1215/mimixbox/internal/applets/debianutils/valid-shell"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if validShell.New() == nil {
		t.Fatal("New() returned nil")
	}
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "shells")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidateShellsValid(t *testing.T) {
	t.Parallel()
	// /bin/sh exists and is executable on every system we run on.
	path := writeFile(t, "/bin/sh\n")

	var out bytes.Buffer
	ok, err := validShell.ValidateShellsForTest(path, &out)
	if err != nil {
		t.Fatalf("validateShells error = %v", err)
	}
	if !ok {
		t.Errorf("ok = false, want true; output = %q", out.String())
	}
	if !strings.Contains(out.String(), "OK: /bin/sh") {
		t.Errorf("output = %q, want an OK line for /bin/sh", out.String())
	}
}

func TestValidateShellsInvalid(t *testing.T) {
	t.Parallel()
	path := writeFile(t, "/no/such/shell\n")

	var out bytes.Buffer
	ok, err := validShell.ValidateShellsForTest(path, &out)
	if err != nil {
		t.Fatalf("validateShells error = %v", err)
	}
	if ok {
		t.Errorf("ok = true, want false for a nonexistent shell")
	}
	if !strings.Contains(out.String(), "NG: /no/such/shell") {
		t.Errorf("output = %q, want an NG line for the missing shell", out.String())
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := validShell.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("help err = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help output missing %q:\n%s", want, out.String())
		}
	}
}
