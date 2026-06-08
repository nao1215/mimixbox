// mimixbox/internal/lib/file_test.go
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
	"testing"

	"github.com/stretchr/testify/assert"
)

// fixtures builds the test files and directories in a per-test temporary
// directory and returns its path. Using t.TempDir() instead of a shared
// /tmp/mimixbox/ut tree (previously created by test/ut/prepareUnitTest.sh) makes
// these tests self-contained — "go test ./..." passes on a clean checkout — and
// parallel-safe, since each test owns its own fixtures.
func fixtures(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	files := map[string]os.FileMode{
		"Executable.txt":    0o755,
		"Writable.txt":      0o644,
		"Readable.txt":      0o444,
		"NonExecutable.txt": 0o644,
		"NonWritable.txt":   0o444,
		"NonReadable.txt":   0o000,
		"AllZero.txt":       0o000,
		".hidden.txt":       0o644,
	}
	for name, mode := range files {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, nil, 0o644); err != nil { //nolint:gosec // fixture file
			t.Fatal(err)
		}
		if err := os.Chmod(p, mode); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.Symlink(filepath.Join(dir, "Executable.txt"), filepath.Join(dir, "symbolic.txt")); err != nil {
		t.Fatal(err)
	}

	noEmpty := filepath.Join(dir, "NoEmptyDir")
	if err := os.MkdirAll(noEmpty, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"aaa.txt", "bbb.txt", "ccc.txt"} {
		if err := os.WriteFile(filepath.Join(noEmpty, name), nil, 0o644); err != nil { //nolint:gosec // fixture file
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "NoWritableDir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(dir, "NoWritableDir"), 0o555); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestIsFile(t *testing.T) {
	t.Parallel()
	dir := fixtures(t)
	assert.Equal(t, true, IsFile(filepath.Join(dir, "Readable.txt")))
	assert.Equal(t, true, IsFile(filepath.Join(dir, "symbolic.txt")))
	assert.Equal(t, false, IsFile(dir))
	assert.Equal(t, true, IsFile(filepath.Join(dir, "AllZero.txt")))
	assert.Equal(t, false, IsFile(filepath.Join(dir, "NoEmptyDir")))
	assert.Equal(t, true, IsFile(filepath.Join(dir, ".hidden.txt")))
	assert.Equal(t, false, IsFile("abcdef"))
}

func TestExists(t *testing.T) {
	t.Parallel()
	dir := fixtures(t)
	assert.Equal(t, true, Exists(filepath.Join(dir, "Readable.txt")))
	assert.Equal(t, true, Exists(filepath.Join(dir, "symbolic.txt")))
	assert.Equal(t, true, Exists(dir))
	assert.Equal(t, true, Exists(filepath.Join(dir, "AllZero.txt")))
	assert.Equal(t, true, Exists("/"))
	assert.Equal(t, false, Exists("abcdef"))
}

func TestIsDir(t *testing.T) {
	t.Parallel()
	dir := fixtures(t)
	assert.Equal(t, false, IsDir(filepath.Join(dir, "Readable.txt")))
	assert.Equal(t, false, IsDir(filepath.Join(dir, "symbolic.txt")))
	assert.Equal(t, true, IsDir(dir))
	assert.Equal(t, false, IsDir(filepath.Join(dir, "AllZero.txt")))
	assert.Equal(t, true, IsDir("/"))
	assert.Equal(t, false, IsDir("abcdef"))
	assert.Equal(t, true, IsDir(filepath.Join(dir, "NoWritableDir")))
}

func TestIsSymlink(t *testing.T) {
	t.Parallel()
	dir := fixtures(t)
	assert.Equal(t, false, IsSymlink(filepath.Join(dir, "Readable.txt")))
	assert.Equal(t, true, IsSymlink(filepath.Join(dir, "symbolic.txt")))
	assert.Equal(t, false, IsSymlink(dir))
	assert.Equal(t, false, IsSymlink(filepath.Join(dir, "AllZero.txt")))
	assert.Equal(t, false, IsSymlink("/"))
	assert.Equal(t, false, IsSymlink("abcdef"))
}

func TestIsZero(t *testing.T) {
	t.Parallel()
	dir := fixtures(t)
	assert.Equal(t, true, IsZero(filepath.Join(dir, "Readable.txt")))
	assert.Equal(t, true, IsZero(filepath.Join(dir, "symbolic.txt")))
	assert.Equal(t, false, IsZero(dir))
	assert.Equal(t, false, IsZero("abcdef"))
	assert.Equal(t, true, IsZero(filepath.Join(dir, "AllZero.txt")))
}

func TestIsReadable(t *testing.T) {
	t.Parallel()
	dir := fixtures(t)
	assert.Equal(t, true, IsReadable(filepath.Join(dir, "Readable.txt")))
	assert.Equal(t, true, IsReadable(filepath.Join(dir, "symbolic.txt")))
	assert.Equal(t, true, IsReadable(dir))
	assert.Equal(t, false, IsReadable(filepath.Join(dir, "NonReadable.txt")))
	assert.Equal(t, false, IsReadable("abcdef"))
	assert.Equal(t, false, IsReadable(filepath.Join(dir, "AllZero.txt")))
}

func TestIsWritable(t *testing.T) {
	t.Parallel()
	dir := fixtures(t)
	assert.Equal(t, true, IsWritable(filepath.Join(dir, "Writable.txt")))
	assert.Equal(t, true, IsWritable(filepath.Join(dir, "symbolic.txt")))
	assert.Equal(t, true, IsWritable(dir))
	assert.Equal(t, false, IsWritable(filepath.Join(dir, "NonWritable.txt")))
	assert.Equal(t, false, IsWritable("abcdef"))
	assert.Equal(t, false, IsWritable(filepath.Join(dir, "AllZero.txt")))
}

func TestIsExecutable(t *testing.T) {
	t.Parallel()
	dir := fixtures(t)
	assert.Equal(t, true, IsExecutable(filepath.Join(dir, "Executable.txt")))
	assert.Equal(t, true, IsExecutable(filepath.Join(dir, "symbolic.txt")))
	assert.Equal(t, true, IsExecutable(dir))
	assert.Equal(t, false, IsExecutable(filepath.Join(dir, "NonExecutable.txt")))
	assert.Equal(t, false, IsExecutable("abcdef"))
	assert.Equal(t, false, IsExecutable(filepath.Join(dir, "AllZero.txt")))
}

func TestIsHiddenFile(t *testing.T) {
	t.Parallel()
	dir := fixtures(t)
	assert.Equal(t, false, IsHiddenFile(filepath.Join(dir, "Executable.txt")))
	assert.Equal(t, true, IsHiddenFile(filepath.Join(dir, ".hidden.txt")))
	assert.Equal(t, false, IsHiddenFile(dir))
	assert.Equal(t, false, IsHiddenFile("abcdef"))
	assert.Equal(t, false, IsHiddenFile(".abcdef"))
}

func TestBaseNameWithoutExt(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Executable", BaseNameWithoutExt("/tmp/mimixbox/ut/Executable.txt"))
	assert.Equal(t, ".hidden", BaseNameWithoutExt("/tmp/mimixbox/ut/.hidden.txt"))
	assert.Equal(t, "file", BaseNameWithoutExt("./file.go"))
	assert.Equal(t, "mimixbox", BaseNameWithoutExt("/tmp/mimixbox"))
	assert.Equal(t, "", BaseNameWithoutExt("/tmp/mimixbox/ut/"))
	assert.Equal(t, "ut", BaseNameWithoutExt("/tmp/mimixbox/ut"))
	assert.Equal(t, "abcdef", BaseNameWithoutExt("abcdef"))
	assert.Equal(t, "", BaseNameWithoutExt(".HiddenDir"))
}
