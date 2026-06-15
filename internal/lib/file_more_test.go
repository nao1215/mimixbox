// mimixbox/internal/lib/file_more_test.go
//
// # Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package mb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCopyAndSize(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Copy(src, dst); err != nil {
		t.Fatalf("Copy error = %v", err)
	}
	got, err := os.ReadFile(dst) //nolint:gosec // reading a file the test wrote
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "0123456789" {
		t.Errorf("copied content = %q", got)
	}
	size, err := Size(src)
	if err != nil || size != 10 {
		t.Errorf("Size = %d, %v, want 10", size, err)
	}
	if _, err := Size("/no/such/file"); err == nil {
		t.Error("Size of a missing file should error")
	}
	if err := Copy("/no/such/src", dst); err == nil {
		t.Error("Copy of a missing source should error")
	}
}

func TestReadFileToStrListAndListToFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "lines.txt")
	if err := ListToFile(p, []string{"one\n", "two\n", "three\n"}); err != nil {
		t.Fatalf("ListToFile error = %v", err)
	}
	lines, err := ReadFileToStrList(p)
	if err != nil {
		t.Fatalf("ReadFileToStrList error = %v", err)
	}
	if len(lines) != 3 || !strings.HasPrefix(lines[0], "one") {
		t.Errorf("ReadFileToStrList = %v", lines)
	}
	if _, err := ReadFileToStrList("/no/such/file"); err == nil {
		t.Error("ReadFileToStrList of a missing file should error")
	}
}

func TestWalk(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, dirs, err := Walk(dir, true)
	if err != nil {
		t.Fatalf("Walk error = %v", err)
	}
	if len(files) < 2 {
		t.Errorf("Walk found %d files, want >= 2", len(files))
	}
	if len(dirs) < 1 {
		t.Errorf("Walk found %d dirs, want >= 1", len(dirs))
	}
}

func TestIsSameFileName(t *testing.T) {
	t.Parallel()
	if !IsSameFileName("/a/b/foo.txt", "/c/d/foo.txt") {
		t.Error("same base names should match")
	}
	if IsSameFileName("/a/foo.txt", "/a/bar.txt") {
		t.Error("different base names should not match")
	}
}

func TestIsNamedPipe(t *testing.T) {
	t.Parallel()
	// A regular file is not a named pipe.
	p := filepath.Join(t.TempDir(), "regular")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if IsNamedPipe(p) {
		t.Error("a regular file is not a named pipe")
	}
}

func TestCopyPreservesModeAndMtime(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("data"), 0o755); err != nil {
		t.Fatal(err)
	}
	mtime := time.Unix(1000000000, 0)
	if err := os.Chtimes(src, mtime, mtime); err != nil {
		t.Fatal(err)
	}

	if err := Copy(src, dst); err != nil {
		t.Fatalf("Copy error = %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("dst mode = %o, want 755 (execute bit must survive)", info.Mode().Perm())
	}
	if !info.ModTime().Equal(mtime) {
		t.Errorf("dst mtime = %v, want %v", info.ModTime(), mtime)
	}
	got, _ := os.ReadFile(dst) //nolint:gosec // reading a file the test wrote
	if string(got) != "data" {
		t.Errorf("dst content = %q", got)
	}
}

func TestCopyTreeRecursivePreservesModes(t *testing.T) {
	t.Parallel()
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

	dst := filepath.Join(dir, "copy")
	if err := CopyTree(src, dst); err != nil {
		t.Fatalf("CopyTree error = %v", err)
	}

	checks := map[string]os.FileMode{
		"sub":          0o700,
		"run.sh":       0o755,
		"sub/note.txt": 0o600,
	}
	for rel, want := range checks {
		info, err := os.Stat(filepath.Join(dst, rel))
		if err != nil {
			t.Errorf("missing %s: %v", rel, err)
			continue
		}
		if info.Mode().Perm() != want {
			t.Errorf("%s mode = %o, want %o", rel, info.Mode().Perm(), want)
		}
	}
	got, _ := os.ReadFile(filepath.Join(dst, "sub", "note.txt")) //nolint:gosec // test-written file
	if string(got) != "hi" {
		t.Errorf("nested content = %q", got)
	}
}

func TestListToFileAtomicPreservesMode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(p, []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := ListToFile(p, []string{"new1\n", "new2\n"}); err != nil {
		t.Fatalf("ListToFile error = %v", err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %o, want 600 (must be preserved across the rewrite)", info.Mode().Perm())
	}
	got, _ := os.ReadFile(p) //nolint:gosec // test-written file
	if string(got) != "new1\nnew2\n" {
		t.Errorf("content = %q", got)
	}
	// No temporary files should be left behind in the directory.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}
