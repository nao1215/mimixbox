// mimixbox/internal/applets/fileutils/mv/mv_internal_test.go
//
// Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package mv

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestCrossDeviceFallbackMovesDirectory forces os.Rename to report a
// cross-device error (EXDEV) so rename() takes its copy+remove fallback, and
// checks that a directory tree is moved with its modes preserved. Before the
// fix the fallback bailed out on directories and dropped file modes.
func TestCrossDeviceFallbackMovesDirectory(t *testing.T) {
	orig := osRename
	osRename = func(_, _ string) error {
		return &os.LinkError{Op: "rename", Err: syscall.EXDEV}
	}
	t.Cleanup(func() { osRename = orig })

	dir := t.TempDir()
	src := filepath.Join(dir, "tree")
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "run.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "note.txt"), []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(dir, "moved")
	if err := rename(src, dest); err != nil {
		t.Fatalf("rename across devices error = %v", err)
	}

	// Source must be gone.
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source still exists after move: %v", err)
	}
	// Destination tree must exist with preserved modes and content.
	checks := map[string]os.FileMode{
		"sub":          0o700,
		"run.sh":       0o755,
		"sub/note.txt": 0o600,
	}
	for rel, want := range checks {
		info, err := os.Stat(filepath.Join(dest, rel))
		if err != nil {
			t.Errorf("missing %s: %v", rel, err)
			continue
		}
		if info.Mode().Perm() != want {
			t.Errorf("%s mode = %o, want %o", rel, info.Mode().Perm(), want)
		}
	}
	got, _ := os.ReadFile(filepath.Join(dest, "sub", "note.txt")) //nolint:gosec // test-written file
	if string(got) != "hi" {
		t.Errorf("moved content = %q", got)
	}
}

// TestCrossDeviceFallbackMovesFile checks the single-file cross-device path
// preserves the file's mode (the execute bit in particular).
func TestCrossDeviceFallbackMovesFile(t *testing.T) {
	orig := osRename
	osRename = func(_, _ string) error {
		return &os.LinkError{Op: "rename", Err: syscall.EXDEV}
	}
	t.Cleanup(func() { osRename = orig })

	dir := t.TempDir()
	src := filepath.Join(dir, "script.sh")
	dest := filepath.Join(dir, "dest.sh")
	if err := os.WriteFile(src, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := rename(src, dest); err != nil {
		t.Fatalf("rename error = %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source still exists: %v", err)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("dest mode = %o, want 755", info.Mode().Perm())
	}
}
