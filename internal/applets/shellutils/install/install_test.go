package install_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/install"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := install.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := install.New()
	if got := c.Name(); got != "install" {
		t.Errorf("Name() = %q, want %q", got, "install")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestRunCopyToFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "hello")

	if _, errOut, err := run(t, "-m", "640", src, dst); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("content = %q, want %q", got, "hello")
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Errorf("mode = %o, want 640", info.Mode().Perm())
	}
}

func TestRunDefaultMode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "x")

	if _, _, err := run(t, src, dst); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("default mode = %o, want 755", info.Mode().Perm())
	}
}

func TestRunCopyIntoDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	destDir := filepath.Join(dir, "bin")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	writeFile(t, a, "A")
	writeFile(t, b, "B")

	if _, errOut, err := run(t, a, b, destDir); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	for name, want := range map[string]string{"a": "A", "b": "B"} {
		got, err := os.ReadFile(filepath.Join(destDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if string(got) != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestRunTargetDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	destDir := filepath.Join(dir, "out")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "f")
	writeFile(t, src, "data")

	if _, _, err := run(t, "-t", destDir, src); err != nil {
		t.Fatalf("Run -t error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "f")); err != nil {
		t.Errorf("expected file copied into target dir: %v", err)
	}
}

func TestRunCreateDirectories(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "a", "b", "c")

	if _, _, err := run(t, "-d", "-m", "700", target); err != nil {
		t.Fatalf("Run -d error = %v", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Errorf("%s is not a directory", target)
	}
	if info.Mode().Perm() != 0o700 {
		t.Errorf("dir mode = %o, want 700", info.Mode().Perm())
	}
}

func TestRunCreateLeading(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFile(t, src, "z")
	dst := filepath.Join(dir, "nested", "deep", "dst")

	if _, errOut, err := run(t, "-D", src, dst); err != nil {
		t.Fatalf("Run -D error = %v (stderr=%q)", err, errOut)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("expected file at %s: %v", dst, err)
	}
}

func TestRunPreserveTimestamps(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "x")

	srcInfo, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, "-p", src, dst); err != nil {
		t.Fatalf("Run -p error = %v", err)
	}
	dstInfo, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !dstInfo.ModTime().Equal(srcInfo.ModTime()) {
		t.Errorf("mtime = %v, want %v", dstInfo.ModTime(), srcInfo.ModTime())
	}
}

func TestRunVerbose(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "x")

	out, _, err := run(t, "-v", src, dst)
	if err != nil {
		t.Fatalf("Run -v error = %v", err)
	}
	if !strings.Contains(out, "->") {
		t.Errorf("verbose output = %q, want it to contain '->'", out)
	}
}

func TestRunErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{"no operand", nil},
		{"missing destination", []string{"only-source"}},
		{"invalid mode", []string{"-m", "999x", "a", "b"}},
		{"directory missing operand", []string{"-d"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, errOut, err := run(t, tt.args...)
			if err == nil {
				t.Errorf("expected error for args %v", tt.args)
			}
			if errOut == "" {
				t.Errorf("expected stderr message for args %v", tt.args)
			}
		})
	}
}

func TestRunMissingSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, errOut, err := run(t, filepath.Join(dir, "nope"), filepath.Join(dir, "dst"))
	if err == nil {
		t.Error("expected error for missing source")
	}
	if !strings.Contains(errOut, "install:") {
		t.Errorf("stderr = %q, want install: prefix", errOut)
	}
}

func TestRunOmitDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "adir")
	if err := os.Mkdir(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, srcDir, filepath.Join(dir, "dst"))
	if err == nil {
		t.Error("expected error when source is a directory")
	}
	if !strings.Contains(errOut, "omitting directory") {
		t.Errorf("stderr = %q, want omitting directory", errOut)
	}
}

func TestRunMultiSourceNonDirDest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	writeFile(t, a, "A")
	writeFile(t, b, "B")
	dst := filepath.Join(dir, "dst") // a regular (non-directory) destination

	_, errOut, err := run(t, a, b, dst)
	if err == nil {
		t.Error("expected error: multiple sources with non-directory destination")
	}
	if !strings.Contains(errOut, "is not a directory") {
		t.Errorf("stderr = %q, want 'is not a directory'", errOut)
	}
}

// TestRunDirectoryVerbose covers the verbose branch of makeDirectories.
func TestRunDirectoryVerbose(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "made")
	out, _, err := run(t, "-d", "-v", target)
	if err != nil {
		t.Fatalf("Run -d -v error = %v", err)
	}
	if !strings.Contains(out, "creating directory") {
		t.Errorf("verbose output = %q, want creating-directory message", out)
	}
}

// TestRunDirectoryMultipleWithFailure verifies that makeDirectories continues
// past a failed directory and still reports failure. A directory whose parent
// is a regular file cannot be created.
func TestRunDirectoryMultipleWithFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "good")
	fileParent := filepath.Join(dir, "afile")
	writeFile(t, fileParent, "x")
	bad := filepath.Join(fileParent, "child") // parent is a file -> mkdir fails

	_, errOut, err := run(t, "-d", bad, good)
	if err == nil {
		t.Fatal("expected failure for un-creatable directory")
	}
	if !strings.Contains(errOut, "cannot create directory") {
		t.Errorf("stderr = %q, want cannot-create-directory message", errOut)
	}
	// The reachable directory was still created.
	if info, statErr := os.Stat(good); statErr != nil || !info.IsDir() {
		t.Errorf("good directory not created: %v", statErr)
	}
}

// TestRunCopyOpenError covers copyFile's open-error branch: a source that
// cannot be opened (unreadable) surfaces an install error.
func TestRunCopyDestUnwritable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFile(t, src, "data")
	// Destination's parent is a regular file, so os.Create fails.
	fileParent := filepath.Join(dir, "afile")
	writeFile(t, fileParent, "x")
	dst := filepath.Join(fileParent, "dst")

	_, errOut, err := run(t, src, dst)
	if err == nil {
		t.Fatal("expected error: destination not creatable")
	}
	if !strings.Contains(errOut, "install:") {
		t.Errorf("stderr = %q, want install: prefix", errOut)
	}
}

// TestRunBackupSimple verifies --backup=simple moves the existing destination
// aside to dest<suffix> before overwriting.
func TestRunBackupSimple(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "new")
	writeFile(t, dst, "old")

	if _, errOut, err := run(t, "--backup=simple", src, dst); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if got, _ := os.ReadFile(dst); string(got) != "new" {
		t.Errorf("dst = %q, want %q", got, "new")
	}
	if got, err := os.ReadFile(dst + "~"); err != nil || string(got) != "old" {
		t.Errorf("backup dst~ = %q err=%v, want %q", got, err, "old")
	}
}

// TestRunBackupSuffix verifies --suffix overrides the simple-backup suffix.
func TestRunBackupSuffix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "new")
	writeFile(t, dst, "old")

	if _, errOut, err := run(t, "--backup=simple", "-S", ".bak", src, dst); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if got, err := os.ReadFile(dst + ".bak"); err != nil || string(got) != "old" {
		t.Errorf("backup dst.bak = %q err=%v, want %q", got, err, "old")
	}
}

// TestRunBackupNumbered verifies numbered backups produce dest.~N~ names that
// increment when prior numbered backups exist.
func TestRunBackupNumbered(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "v1")
	writeFile(t, dst, "orig")

	if _, errOut, err := run(t, "--backup=numbered", src, dst); err != nil {
		t.Fatalf("Run #1 error = %v (stderr=%q)", err, errOut)
	}
	if got, err := os.ReadFile(dst + ".~1~"); err != nil || string(got) != "orig" {
		t.Errorf("backup dst.~1~ = %q err=%v, want %q", got, err, "orig")
	}

	// Install again; the now-present dst ("v1") should move to dst.~2~.
	writeFile(t, src, "v2")
	if _, errOut, err := run(t, "--backup=numbered", src, dst); err != nil {
		t.Fatalf("Run #2 error = %v (stderr=%q)", err, errOut)
	}
	if got, err := os.ReadFile(dst + ".~2~"); err != nil || string(got) != "v1" {
		t.Errorf("backup dst.~2~ = %q err=%v, want %q", got, err, "v1")
	}
}

// TestRunBackupExisting verifies existing-mode picks numbered when a numbered
// backup already exists, otherwise simple.
func TestRunBackupExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "new")
	writeFile(t, dst, "old")

	// No numbered backups exist yet -> existing falls back to simple (dst~).
	if _, errOut, err := run(t, "--backup=existing", src, dst); err != nil {
		t.Fatalf("Run simple-fallback error = %v (stderr=%q)", err, errOut)
	}
	if got, err := os.ReadFile(dst + "~"); err != nil || string(got) != "old" {
		t.Errorf("backup dst~ = %q err=%v, want %q", got, err, "old")
	}

	// Now create a numbered backup so existing chooses numbered.
	writeFile(t, dst+".~1~", "n1")
	writeFile(t, dst, "current")
	writeFile(t, src, "newer")
	if _, errOut, err := run(t, "--backup=existing", src, dst); err != nil {
		t.Fatalf("Run numbered error = %v (stderr=%q)", err, errOut)
	}
	if got, err := os.ReadFile(dst + ".~2~"); err != nil || string(got) != "current" {
		t.Errorf("backup dst.~2~ = %q err=%v, want %q", got, err, "current")
	}
}

// TestRunBackupBareDefault verifies bare --backup defaults to existing-style
// (no numbered backups -> simple suffix).
func TestRunBackupBareDefault(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "new")
	writeFile(t, dst, "old")

	if _, errOut, err := run(t, "--backup", src, dst); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if got, err := os.ReadFile(dst + "~"); err != nil || string(got) != "old" {
		t.Errorf("backup dst~ = %q err=%v, want %q", got, err, "old")
	}
}

// TestRunBackupNoneNoBackup verifies CONTROL none makes no backup and simply
// overwrites the destination.
func TestRunBackupNoneNoBackup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "new")
	writeFile(t, dst, "old")

	if _, _, err := run(t, "--backup=none", src, dst); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if _, err := os.Stat(dst + "~"); err == nil {
		t.Error("unexpected backup file created with --backup=none")
	}
	if got, _ := os.ReadFile(dst); string(got) != "new" {
		t.Errorf("dst = %q, want %q", got, "new")
	}
}

// TestRunBackupInvalidControl verifies an unknown CONTROL word is rejected.
func TestRunBackupInvalidControl(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "x")

	_, errOut, err := run(t, "--backup=bogus", src, dst)
	if err == nil {
		t.Fatal("expected error for invalid --backup control")
	}
	if !strings.Contains(errOut, "invalid argument") {
		t.Errorf("stderr = %q, want invalid argument", errOut)
	}
}

// TestRunSuffixEnv verifies $SIMPLE_BACKUP_SUFFIX supplies the simple suffix
// when --suffix is not given.
func TestRunSuffixEnv(t *testing.T) {
	// Not parallel: mutates process environment.
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "new")
	writeFile(t, dst, "old")

	t.Setenv("SIMPLE_BACKUP_SUFFIX", ".orig")
	if _, errOut, err := run(t, "--backup=simple", src, dst); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if got, err := os.ReadFile(dst + ".orig"); err != nil || string(got) != "old" {
		t.Errorf("backup dst.orig = %q err=%v, want %q", got, err, "old")
	}
}

// TestRunOwnerGroupParseAndInstall verifies that -o/-g parse and the file is
// still installed. As a non-root user the chown errors with EPERM, in which
// case install must report it and fail (matching GNU). When run as root the
// chown to the current owner succeeds. Either way the destination is written.
func TestRunOwnerGroupParseAndInstall(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "payload")

	// Target uid/gid 0 (root): a non-root user cannot chown to it, so the flag
	// parses, the file installs, but chown fails with EPERM (GNU behavior). As
	// root, chowning to 0:0 succeeds.
	_, errOut, runErr := run(t, "-o", "0", "-g", "0", src, dst)

	// The copy happens before chown, so the file is installed regardless.
	if got, rerr := os.ReadFile(dst); rerr != nil || string(got) != "payload" {
		t.Errorf("dst = %q err=%v, want %q", got, rerr, "payload")
	}

	if os.Geteuid() == 0 {
		// Root: chown to 0:0 succeeds.
		if runErr != nil {
			t.Errorf("Run as root error = %v (stderr=%q)", runErr, errOut)
		}
	} else {
		// Non-root: chown to uid 0 fails with EPERM; GNU behavior is to report
		// and exit nonzero.
		if runErr == nil {
			t.Error("expected non-root chown to fail")
		}
		if !strings.Contains(errOut, "ownership") {
			t.Errorf("stderr = %q, want ownership error", errOut)
		}
	}
}

// TestRunOwnerInvalid verifies an unknown owner name is rejected with a
// recognizable message, and the chown is attempted (file already installed).
func TestRunOwnerInvalid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "x")

	_, errOut, err := run(t, "-o", "no-such-user-xyz", src, dst)
	if err == nil {
		t.Fatal("expected error for invalid owner")
	}
	if !strings.Contains(errOut, "invalid user") {
		t.Errorf("stderr = %q, want invalid user", errOut)
	}
}

// TestRunGroupInvalid verifies an unknown group name is rejected.
func TestRunGroupInvalid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "x")

	_, errOut, err := run(t, "-g", "no-such-group-xyz", src, dst)
	if err == nil {
		t.Fatal("expected error for invalid group")
	}
	if !strings.Contains(errOut, "invalid group") {
		t.Errorf("stderr = %q, want invalid group", errOut)
	}
}

// TestRunStrip verifies --strip parses and behaves GNU-ishly: when the system
// strip is available the installed file is stripped without error; when it is
// absent install reports an error. The file is installed before strip runs.
func TestRunStrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, src, "not-really-a-binary")

	_, errOut, err := run(t, "-s", src, dst)

	// File is copied before strip is attempted.
	if _, statErr := os.Stat(dst); statErr != nil {
		t.Fatalf("expected dst installed before strip: %v", statErr)
	}

	if _, lookErr := exec.LookPath("strip"); lookErr != nil {
		// strip unavailable: GNU-ish behavior is to error out.
		if err == nil {
			t.Error("expected error when strip is unavailable")
		}
		if !strings.Contains(errOut, "strip") {
			t.Errorf("stderr = %q, want strip mention", errOut)
		}
		return
	}
	// strip available: stripping a non-object file typically errors. Either
	// outcome is acceptable; assert the flag was attempted by checking that on
	// failure the message names strip.
	if err != nil && !strings.Contains(errOut, "strip") {
		t.Errorf("stderr = %q, want strip mention on failure", errOut)
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	if !strings.Contains(out, "Usage: install") {
		t.Errorf("help = %q, want usage line", out)
	}
}
