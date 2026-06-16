package mktemp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNameAndSynopsis covers the metadata accessors.
func TestNameAndSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "mktemp" {
		t.Errorf("Name() = %q, want %q", c.Name(), "mktemp")
	}
	if c.Synopsis() != "Create a temporary file or directory" {
		t.Errorf("Synopsis() = %q", c.Synopsis())
	}
}

// TestTempDir returns $TMPDIR when set and /tmp otherwise.
func TestTempDir(t *testing.T) {
	t.Setenv("TMPDIR", "/custom/tmp")
	if got := tempDir(); got != "/custom/tmp" {
		t.Errorf("tempDir() = %q, want %q", got, "/custom/tmp")
	}
	if err := os.Unsetenv("TMPDIR"); err != nil {
		t.Fatal(err)
	}
	if got := tempDir(); got != "/tmp" {
		t.Errorf("tempDir() = %q, want %q", got, "/tmp")
	}
}

// TestResolve covers the plain form, the tmpdir/-t component form, and the
// rejection of a separator in a component-only template.
func TestResolve(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		template    string
		dir         string
		tmpdirGiven bool
		tFlag       bool
		want        string
		wantErr     bool
	}{
		{"plain", "tmp.XXXX", "/tmp", false, false, "tmp.XXXX", false},
		{"tmpdir given", "build.XXXX", "/var/tmp", true, false, filepath.Join("/var/tmp", "build.XXXX"), false},
		{"t flag", "job.XXXX", "/var/tmp", false, true, filepath.Join("/var/tmp", "job.XXXX"), false},
		{"separator rejected", "a/b.XXXX", "/tmp", true, false, "", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := resolve(tt.template, tt.dir, tt.tmpdirGiven, tt.tFlag)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolve err = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("resolve = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSplitTemplate locates the trailing X run and enforces the minimum count.
func TestSplitTemplate(t *testing.T) {
	t.Parallel()
	prefix, suffix, xs, err := splitTemplate("/tmp/tmp.XXXXXX")
	if err != nil {
		t.Fatalf("splitTemplate err = %v", err)
	}
	if prefix != "/tmp/tmp." || suffix != "" || xs != 6 {
		t.Errorf("splitTemplate = (%q, %q, %d)", prefix, suffix, xs)
	}
	if _, _, _, err := splitTemplate("/tmp/no-x"); err == nil {
		t.Error("expected error for too few X's")
	}
}

// TestGenerateDryRun computes a name without creating anything.
func TestGenerateDryRun(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tmpl := filepath.Join(dir, "tmp.XXXXXX")
	name, err := generate(tmpl, false, true)
	if err != nil {
		t.Fatalf("generate dry-run err = %v", err)
	}
	if !strings.HasPrefix(name, filepath.Join(dir, "tmp.")) {
		t.Errorf("name = %q, want prefix %q", name, filepath.Join(dir, "tmp."))
	}
	if len(filepath.Base(name)) != len("tmp.")+6 {
		t.Errorf("name base = %q, want 6 random chars after prefix", filepath.Base(name))
	}
	if _, statErr := os.Stat(name); !os.IsNotExist(statErr) {
		t.Errorf("dry-run must not create %s", name)
	}
}

// TestGenerateCreatesFileAndDir checks both creation paths place a real entry
// matching the template prefix.
func TestGenerateCreatesFileAndDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	fileName, err := generate(filepath.Join(dir, "f.XXXXXX"), false, false)
	if err != nil {
		t.Fatalf("generate file err = %v", err)
	}
	if info, statErr := os.Stat(fileName); statErr != nil || info.IsDir() {
		t.Errorf("expected a regular file at %s (err=%v)", fileName, statErr)
	}

	dirName, err := generate(filepath.Join(dir, "d.XXXXXX"), true, false)
	if err != nil {
		t.Fatalf("generate dir err = %v", err)
	}
	if info, statErr := os.Stat(dirName); statErr != nil || !info.IsDir() {
		t.Errorf("expected a directory at %s (err=%v)", dirName, statErr)
	}
}

// TestGenerateBadTemplate surfaces the splitTemplate error through generate.
func TestGenerateBadTemplate(t *testing.T) {
	t.Parallel()
	if _, err := generate("/tmp/noXhere", false, true); err == nil {
		t.Error("expected error for template without enough X's")
	}
}
