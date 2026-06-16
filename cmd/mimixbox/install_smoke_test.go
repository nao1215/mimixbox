package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestInstallRemoveSmoke (GitHub issue #790) is an end-to-end smoke test of the
// install / dispatch / remove lifecycle against a real mimixbox binary in a
// temp directory:
//
//  1. --full-install populates an install dir with applet symlinks.
//  2. A representative symlink (cat) resolves back to the mimixbox binary, and
//     invoking it dispatches the real applet (cat --version banner).
//  3. --remove deletes the symlinks while leaving the mimixbox binary in place.
func TestInstallRemoveSmoke(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not available; skipping install smoke test")
	}

	// Build the mimixbox binary once into a temp dir from this package's source.
	binDir := t.TempDir()
	bin := filepath.Join(binDir, "mimixbox")
	pkgDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	build := exec.Command(goBin, "build", "-o", bin, ".")
	build.Dir = pkgDir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	wantBin, err := filepath.EvalSymlinks(bin)
	if err != nil {
		t.Fatalf("eval mimixbox binary path: %v", err)
	}

	installDir := t.TempDir()

	// Install every applet as a symlink in installDir.
	if out, err := exec.Command(bin, "--full-install", installDir).CombinedOutput(); err != nil {
		t.Fatalf("--full-install failed: %v\n%s", err, out)
	}

	// A representative symlink must exist and resolve to the mimixbox binary.
	catLink := filepath.Join(installDir, "cat")
	info, err := os.Lstat(catLink)
	if err != nil {
		t.Fatalf("expected symlink %s to exist: %v", catLink, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is not a symlink (mode %v)", catLink, info.Mode())
	}
	resolved, err := filepath.EvalSymlinks(catLink)
	if err != nil {
		t.Fatalf("resolve %s: %v", catLink, err)
	}
	if resolved != wantBin {
		t.Errorf("%s resolves to %q, want %q", catLink, resolved, wantBin)
	}

	// Invoking the symlink must dispatch the real cat applet.
	out, err := exec.Command(catLink, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("%s --version failed: %v\n%s", catLink, err, out)
	}
	if !strings.Contains(string(out), "cat (mimixbox)") {
		t.Errorf("%s --version output %q does not contain %q", catLink, out, "cat (mimixbox)")
	}

	// Remove the symlinks MimixBox created.
	if out, err := exec.Command(bin, "--remove", installDir).CombinedOutput(); err != nil {
		t.Fatalf("--remove failed: %v\n%s", err, out)
	}

	// The symlink must be gone...
	if _, err := os.Lstat(catLink); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, lstat err = %v", catLink, err)
	}
	// ...while the mimixbox binary itself remains.
	if _, err := os.Stat(bin); err != nil {
		t.Errorf("mimixbox binary should still exist after --remove: %v", err)
	}
}
