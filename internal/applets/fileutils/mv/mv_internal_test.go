// mimixbox/internal/applets/fileutils/mv/mv_internal_test.go
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
package mv

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
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

// TestValidArgs checks the mutually exclusive option combinations are rejected
// and that benign combinations pass.
func TestValidArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		opts    options
		wantErr bool
	}{
		{"noclobber+backup", options{noClobber: true, backup: true}, true},
		{"noclobber+force", options{noClobber: true, force: true}, true},
		{"force+interactive", options{force: true, interactive: true}, true},
		{"noclobber+interactive", options{noClobber: true, interactive: true}, true},
		{"backup+interactive ok", options{backup: true, interactive: true}, false},
		{"none", options{}, false},
		{"verbose only", options{verbose: true}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validArgs(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("validArgs(%+v) error = %v, wantErr %v", tt.opts, err, tt.wantErr)
			}
		})
	}
}

// TestQuestion drives the interactive prompt's answer parsing without a real
// terminal by feeding answers through stdin.
func TestQuestion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"y", "y\n", true},
		{"yes upper", "YES\n", true},
		{"n", "n\n", false},
		{"empty line", "\n", false},
		{"garbage", "maybe\n", false},
		{"eof", "", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := &bytes.Buffer{}
			stdio := command.IO{In: strings.NewReader(tt.input), Out: out, Err: &bytes.Buffer{}}
			if got := question(stdio, "Overwrite x"); got != tt.want {
				t.Errorf("question(%q) = %v, want %v", tt.input, got, tt.want)
			}
			if !strings.Contains(out.String(), "Overwrite x [Y/n]") {
				t.Errorf("prompt = %q, want it to contain the question", out.String())
			}
		})
	}
}

// TestForceMoveOverwrites confirms forceMove replaces an existing destination
// file with the source content.
func TestForceMoveOverwrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dest := filepath.Join(dir, "dest.txt")
	if err := os.WriteFile(src, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := forceMove(src, dest, options{force: true}); err != nil {
		t.Fatalf("forceMove error = %v", err)
	}
	got, _ := os.ReadFile(dest) //nolint:gosec // test-written file
	if string(got) != "new" {
		t.Errorf("dest content = %q, want %q", got, "new")
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source still exists: %v", err)
	}
}

// TestInteractiveMoveYesOverwrites checks that answering yes lets the move
// proceed and overwrite the destination.
func TestInteractiveMoveYesOverwrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "f.txt")
	destDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(destDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(destDir, "f.txt")
	if err := os.WriteFile(existing, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdio := command.IO{In: strings.NewReader("y\n"), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := interactiveMove(stdio, src, destDir, options{interactive: true}); err != nil {
		t.Fatalf("interactiveMove error = %v", err)
	}
	got, _ := os.ReadFile(existing) //nolint:gosec // test-written file
	if string(got) != "new" {
		t.Errorf("dest content = %q, want overwrite to %q", got, "new")
	}
}

// TestInteractiveMoveNoKeepsDestination checks that answering no leaves the
// destination untouched and does not move the source.
func TestInteractiveMoveNoKeepsDestination(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "f.txt")
	destDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(destDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(destDir, "f.txt")
	if err := os.WriteFile(existing, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdio := command.IO{In: strings.NewReader("n\n"), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := interactiveMove(stdio, src, destDir, options{interactive: true}); err != nil {
		t.Fatalf("interactiveMove error = %v", err)
	}
	got, _ := os.ReadFile(existing) //nolint:gosec // test-written file
	if string(got) != "old" {
		t.Errorf("dest content = %q, want %q (must not overwrite)", got, "old")
	}
	if _, err := os.Stat(src); err != nil {
		t.Errorf("source should remain after declined overwrite: %v", err)
	}
}

// TestDecideBackupFileName checks the backup suffix logic, including the
// recursive case where the first backup name is already taken.
func TestDecideBackupFileName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	first := decideBackupFileName(target)
	if first != target+"~" {
		t.Errorf("backup name = %q, want %q", first, target+"~")
	}

	// Create the first backup so the next call must recurse to a further name.
	if err := os.WriteFile(first, []byte("y"), 0o600); err != nil {
		t.Fatal(err)
	}
	second := decideBackupFileName(target)
	if second != target+"~~" {
		t.Errorf("recursive backup name = %q, want %q", second, target+"~~")
	}
}

// TestNoclobberMoveSkipsSameName verifies that moving a file onto a directory
// that already holds a same-named file is a no-op under -n.
func TestNoclobberMoveSkipsSameName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "f.txt")
	destDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(destDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(destDir, "f.txt")
	if err := os.WriteFile(existing, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := noclobberMove(src, destDir); err != nil {
		t.Fatalf("noclobberMove error = %v", err)
	}
	got, _ := os.ReadFile(existing) //nolint:gosec // test-written file
	if string(got) != "old" {
		t.Errorf("dest content = %q, want %q (no clobber)", got, "old")
	}
	// Source must still be present because nothing was moved.
	if _, err := os.Stat(src); err != nil {
		t.Errorf("source should remain: %v", err)
	}
}

// TestIsSameNameFileOrDir exercises the branch matrix of isSameNameFileOrDir.
func TestIsSameNameFileOrDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	srcDir := filepath.Join(dir, "a", "data")
	destDir := filepath.Join(dir, "b", "data")
	if err := os.MkdirAll(srcDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if !isSameNameFileOrDir(srcDir, destDir) {
		t.Error("same-named dirs should report true")
	}

	srcFile := filepath.Join(dir, "x.txt")
	destFile := filepath.Join(dir, "c", "x.txt")
	if err := os.Mkdir(filepath.Join(dir, "c"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcFile, []byte("1"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destFile, []byte("2"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !isSameNameFileOrDir(srcFile, destFile) {
		t.Error("same-named files should report true")
	}

	// File into a directory that already contains a same-named entry.
	intoDir := filepath.Join(dir, "c")
	if !isSameNameFileOrDir(srcFile, intoDir) {
		t.Error("file into dir holding same name should report true")
	}

	// Differently named file into a directory without that name.
	other := filepath.Join(dir, "y.txt")
	if err := os.WriteFile(other, []byte("3"), 0o600); err != nil {
		t.Fatal(err)
	}
	if isSameNameFileOrDir(other, filepath.Join(dir, "a")) {
		t.Error("absent name in dest dir should report false")
	}
}
