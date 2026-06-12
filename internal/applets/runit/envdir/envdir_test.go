package envdir

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func envDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestSetsEnvForProgram(t *testing.T) {
	dir := envDir(t, map[string]string{"FOO": "bar\n", "MULTI": "first\nsecond\n"})
	out, err := run(t, dir, "sh", "-c", `printf '%s|%s' "$FOO" "$MULTI"`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "bar|first" {
		t.Errorf("env = %q, want bar|first", out)
	}
}

func TestEmptyFileRemovesVar(t *testing.T) {
	dir := envDir(t, map[string]string{"REMOVED": ""})
	t.Setenv("REMOVED", "present")
	out, err := run(t, dir, "sh", "-c", `printf '[%s]' "$REMOVED"`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "[]" {
		t.Errorf("empty file should remove the var, got %q", out)
	}
}

func TestTrailingWhitespaceTrimmed(t *testing.T) {
	dir := envDir(t, map[string]string{"V": "value   \t\n"})
	got, _ := applyDir([]string{}, dir)
	found := ""
	for _, kv := range got {
		if strings.HasPrefix(kv, "V=") {
			found = kv
		}
	}
	if found != "V=value" {
		t.Errorf("trailing whitespace not trimmed: %q", found)
	}
}

func TestExitCodePropagates(t *testing.T) {
	dir := envDir(t, map[string]string{"X": "1"})
	_, err := run(t, dir, "sh", "-c", "exit 4")
	ee, ok := err.(*command.ExitError)
	if !ok || ee.Code != 4 {
		t.Errorf("err = %v, want exit 4", err)
	}
}

func TestErrors(t *testing.T) {
	if _, err := run(t, "/only-one-arg"); err == nil {
		t.Errorf("too few args should fail")
	}
	if _, err := run(t, "/no/such/dir", "true"); err == nil {
		t.Errorf("a missing directory should fail")
	}
}

func TestSkipsFilenamesWithEquals(t *testing.T) {
	dir := envDir(t, map[string]string{"A=B": "content", "OK": "value"})
	got, err := applyDir([]string{}, dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range got {
		if strings.HasPrefix(kv, "A=") {
			t.Errorf("a filename containing '=' must be skipped, got %q", kv)
		}
	}
	// The well-formed file is still applied.
	var hasOK bool
	for _, kv := range got {
		if kv == "OK=value" {
			hasOK = true
		}
	}
	if !hasOK {
		t.Errorf("the valid file should still be applied: %v", got)
	}
}
