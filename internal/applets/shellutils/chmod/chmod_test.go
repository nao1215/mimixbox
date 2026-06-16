package chmod_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/chmod"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := chmod.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestMetadata(t *testing.T) {
	c := chmod.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Name() != "chmod" {
		t.Errorf("Name() = %q, want chmod", c.Name())
	}
	if c.Synopsis() != "Change file mode bits" {
		t.Errorf("Synopsis() = %q", c.Synopsis())
	}
}

// TestApplyModeOctal checks the numeric form.
func TestApplyModeOctal(t *testing.T) {
	tests := []struct {
		name string
		cur  os.FileMode
		mode string
		want os.FileMode
	}{
		{"644", 0o600, "644", 0o644},
		{"0644 leading zero", 0o600, "0644", 0o644},
		{"755", 0o644, "755", 0o755},
		{"600", 0o777, "600", 0o600},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := chmod.ApplyModeForTest(tt.cur, tt.mode, false)
			if err != nil {
				t.Fatalf("ApplyMode error: %v", err)
			}
			if got.Perm() != tt.want {
				t.Errorf("ApplyMode(%v,%q) = %o, want %o", tt.cur, tt.mode, got.Perm(), tt.want)
			}
		})
	}
}

// TestApplyModeSymbolic checks symbolic clauses against known starting modes.
func TestApplyModeSymbolic(t *testing.T) {
	tests := []struct {
		name  string
		cur   os.FileMode
		mode  string
		isDir bool
		want  os.FileMode
	}{
		{"u+x adds owner execute", 0o644, "u+x", false, 0o744},
		{"go-w removes group/other write", 0o666, "go-w", false, 0o644},
		{"a=r sets all to read only", 0o755, "a=r", false, 0o444},
		{"+x adds execute to all", 0o644, "+x", false, 0o755},
		{"u-x removes owner execute", 0o755, "u-x", false, 0o655},
		{"comma list", 0o600, "u+x,g+r", false, 0o740},
		{"X on dir adds execute", 0o644, "+X", true, 0o755},
		{"X on plain file no execute", 0o644, "+X", false, 0o644},
		{"X on already-exec file", 0o744, "g+X", false, 0o754},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := chmod.ApplyModeForTest(tt.cur, tt.mode, tt.isDir)
			if err != nil {
				t.Fatalf("ApplyMode error: %v", err)
			}
			if got.Perm() != tt.want {
				t.Errorf("ApplyMode(%o,%q) = %o, want %o", tt.cur.Perm(), tt.mode, got.Perm(), tt.want)
			}
		})
	}
}

func TestApplyModeSpecialBits(t *testing.T) {
	got, err := chmod.ApplyModeForTest(0o755, "u+s", false)
	if err != nil {
		t.Fatal(err)
	}
	if got&os.ModeSetuid == 0 {
		t.Errorf("u+s did not set setuid: %v", got)
	}
	got, err = chmod.ApplyModeForTest(0o755, "+t", true)
	if err != nil {
		t.Fatal(err)
	}
	if got&os.ModeSticky == 0 {
		t.Errorf("+t did not set sticky: %v", got)
	}
}

func TestApplyModeInvalid(t *testing.T) {
	if _, err := chmod.ApplyModeForTest(0o644, "u?x", false); err == nil {
		t.Error("expected error for invalid mode")
	}
	if _, err := chmod.ApplyModeForTest(0o644, "", false); err == nil {
		t.Error("expected error for empty mode")
	}
}

// TestRunOctal exercises Run end-to-end with an octal mode.
func TestRunOctal(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, "644", file)
	if err != nil {
		t.Fatalf("Run error = %v, stderr = %q", err, errOut)
	}
	st, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o644 {
		t.Errorf("perm = %o, want 0644", st.Mode().Perm())
	}
}

// TestRunSymbolic exercises Run end-to-end with symbolic clauses.
func TestRunSymbolic(t *testing.T) {
	tests := []struct {
		mode  string
		start os.FileMode
		want  os.FileMode
	}{
		{"u+x", 0o644, 0o744},
		{"go-w", 0o666, 0o644},
		{"a=r", 0o755, 0o444},
		{"+x", 0o644, 0o755},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			dir := t.TempDir()
			file := filepath.Join(dir, "f.txt")
			if err := os.WriteFile(file, []byte("x\n"), tt.start); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(file, tt.start); err != nil {
				t.Fatal(err)
			}
			_, errOut, err := run(t, tt.mode, file)
			if err != nil {
				t.Fatalf("Run error = %v, stderr = %q", err, errOut)
			}
			st, err := os.Stat(file)
			if err != nil {
				t.Fatal(err)
			}
			if st.Mode().Perm() != tt.want {
				t.Errorf("mode %q: perm = %o, want %o", tt.mode, st.Mode().Perm(), tt.want)
			}
		})
	}
}

// TestRunRecursive applies a mode to a directory tree.
func TestRunRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	f1 := filepath.Join(dir, "a.txt")
	f2 := filepath.Join(sub, "b.txt")
	for _, f := range []string{f1, f2} {
		if err := os.WriteFile(f, []byte("x\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// Use a mode that keeps directories traversable (0755) so the post-change
	// stats below, and TempDir cleanup, can still descend the tree.
	_, errOut, err := run(t, "-R", "755", dir)
	if err != nil {
		t.Fatalf("Run error = %v, stderr = %q", err, errOut)
	}
	for _, p := range []string{dir, sub, f1, f2} {
		st, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if st.Mode().Perm() != 0o755 {
			t.Errorf("%s perm = %o, want 0755", p, st.Mode().Perm())
		}
	}
}

// TestRunMissingFile reports an error and non-zero exit for a missing file.
func TestRunMissingFile(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.txt")

	_, errOut, err := run(t, "644", missing)
	if err == nil {
		t.Fatal("expected error (exit 1) for missing file")
	}
	if !strings.Contains(errOut, "cannot access") {
		t.Errorf("stderr = %q, want to contain 'cannot access'", errOut)
	}
}

// TestRunMissingOperand checks the missing operand error.
func TestRunMissingOperand(t *testing.T) {
	if _, errOut, err := run(t); err == nil {
		t.Errorf("expected error for no operands, stderr=%q", errOut)
	}
	if _, errOut, err := run(t, "644"); err == nil {
		t.Errorf("expected error for missing file operand, stderr=%q", errOut)
	}
}

// TestRunVerbose checks that -v emits a diagnostic.
func TestRunVerbose(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "-v", "755", file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "mode of") {
		t.Errorf("verbose stdout = %q, want diagnostic", out)
	}
}

// TestHelpSections verifies that --help renders both the Examples and the
// Exit status sections supplied through WithHelp.
func TestHelpSections(t *testing.T) {
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help missing %q section:\n%s", want, out)
		}
	}
}
