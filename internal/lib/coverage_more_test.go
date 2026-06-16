// mimixbox/internal/lib/coverage_more_test.go
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
	"bytes"
	"crypto/md5"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// crypto.go
// ---------------------------------------------------------------------------

func TestCompareChecksumOK(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	target := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(target, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	sum, err := CalcChecksum(md5.New(), target)
	if err != nil {
		t.Fatal(err)
	}

	// Coreutils-style checksum file uses two spaces between digest and path.
	sumFile := filepath.Join(dir, "sums.md5")
	if err := os.WriteFile(sumFile, []byte(sum+"  "+target+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := CompareChecksum(&out, md5.New(), []string{sumFile}); err != nil {
		t.Fatalf("CompareChecksum error = %v", err)
	}
	if got := out.String(); !strings.Contains(got, target+": OK") {
		t.Errorf("expected OK line, got %q", got)
	}
}

func TestCompareChecksumFail(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	target := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(target, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Deliberately wrong digest -> Fail line.
	sumFile := filepath.Join(dir, "sums.md5")
	if err := os.WriteFile(sumFile, []byte("00000000000000000000000000000000  "+target+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := CompareChecksum(&out, md5.New(), []string{sumFile}); err != nil {
		t.Fatalf("CompareChecksum error = %v", err)
	}
	if got := out.String(); !strings.Contains(got, target+": Fail") {
		t.Errorf("expected Fail line, got %q", got)
	}
}

func TestCompareChecksumMissingChecksumFile(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	err := CompareChecksum(&out, md5.New(), []string{filepath.Join(t.TempDir(), "nope.md5")})
	if err == nil {
		t.Fatal("expected error for missing checksum file, got nil")
	}
}

func TestCompareChecksumWrongFormat(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sumFile := filepath.Join(dir, "bad.md5")
	// Single space, not the required double space, so Split yields one field.
	if err := os.WriteFile(sumFile, []byte("deadbeef file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := CompareChecksum(&out, md5.New(), []string{sumFile})
	if err == nil || !strings.Contains(err.Error(), "wrong checksum format") {
		t.Fatalf("expected wrong checksum format error, got %v", err)
	}
}

func TestCompareChecksumMissingTargetFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sumFile := filepath.Join(dir, "sums.md5")
	missing := filepath.Join(dir, "ghost.txt")
	if err := os.WriteFile(sumFile, []byte("deadbeef  "+missing+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := CompareChecksum(&out, md5.New(), []string{sumFile}); err == nil {
		t.Fatal("expected error opening missing target file, got nil")
	}
}

func TestChecksumOutput(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	r := strings.NewReader("hello")
	if err := ChecksumOutput(&out, md5.New(), r, "data.txt"); err != nil {
		t.Fatalf("ChecksumOutput error = %v", err)
	}
	// md5("hello") = 5d41402abc4b2a76b9719d911017c592
	want := "5d41402abc4b2a76b9719d911017c592  data.txt\n"
	if out.String() != want {
		t.Errorf("ChecksumOutput = %q, want %q", out.String(), want)
	}
}

func TestPrintChecksumsMixed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	good := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(good, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing.txt")
	subdir := filepath.Join(dir, "adir")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errw bytes.Buffer
	status, err := PrintChecksums(&out, &errw, "md5sum", md5.New(), []string{good, missing, subdir})
	if err != nil {
		t.Fatalf("PrintChecksums error = %v", err)
	}
	if status != 1 {
		t.Errorf("status = %d, want 1 (some entries failed)", status)
	}
	if !strings.Contains(out.String(), "5d41402abc4b2a76b9719d911017c592  "+good) {
		t.Errorf("missing digest for good file, out=%q", out.String())
	}
	if !strings.Contains(errw.String(), "No such file or directory") {
		t.Errorf("missing 'No such file' diagnostic, errw=%q", errw.String())
	}
	if !strings.Contains(errw.String(), "It is directory") {
		t.Errorf("missing 'It is directory' diagnostic, errw=%q", errw.String())
	}
}

func TestPrintChecksumsAllGood(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errw bytes.Buffer
	status, err := PrintChecksums(&out, &errw, "md5sum", md5.New(), []string{f})
	if err != nil {
		t.Fatalf("PrintChecksums error = %v", err)
	}
	if status != 0 {
		t.Errorf("status = %d, want 0", status)
	}
	if errw.Len() != 0 {
		t.Errorf("expected no diagnostics, got %q", errw.String())
	}
}

func TestCalcChecksumMissingFile(t *testing.T) {
	t.Parallel()
	_, err := CalcChecksum(md5.New(), filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// ---------------------------------------------------------------------------
// file.go
// ---------------------------------------------------------------------------

func TestCopyMissingSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	err := Copy(filepath.Join(dir, "nope"), filepath.Join(dir, "dst"))
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestCopyDestInMissingDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Destination directory does not exist -> OpenFile fails.
	err := Copy(src, filepath.Join(dir, "no", "such", "dst"))
	if err == nil {
		t.Fatal("expected error opening dest in missing dir, got nil")
	}
}

func TestCopyTreeMissingSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := CopyTree(filepath.Join(dir, "nope"), filepath.Join(dir, "dst")); err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestCopyTreeSingleFileFallsBackToCopy(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CopyTree(src, dst); err != nil {
		t.Fatalf("CopyTree error = %v", err)
	}
	got, err := os.ReadFile(dst) //nolint:gosec // reading a file the test wrote
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "content" {
		t.Errorf("CopyTree single file = %q, want %q", got, "content")
	}
}

func TestReadFileToStrListMissing(t *testing.T) {
	t.Parallel()
	if _, err := ReadFileToStrList(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestListToFileTempDirCreationError(t *testing.T) {
	t.Parallel()
	// Parent directory does not exist, so CreateTemp fails.
	err := ListToFile(filepath.Join(t.TempDir(), "no-such-dir", "out.txt"), []string{"a"})
	if err == nil {
		t.Fatal("expected error creating temp in missing dir, got nil")
	}
}

func TestListToFileOverwritePreservesMode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("old\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := ListToFile(path, []string{"line1\n", "line2\n"}); err != nil {
		t.Fatalf("ListToFile error = %v", err)
	}
	got, err := os.ReadFile(path) //nolint:gosec // reading a file the test wrote
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "line1\nline2\n" {
		t.Errorf("ListToFile content = %q", got)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o640 {
			t.Errorf("mode = %o, want 0640 (preserved)", info.Mode().Perm())
		}
	}
}

func TestWalkMissingDirReturnsError(t *testing.T) {
	t.Parallel()
	_, _, err := Walk(filepath.Join(t.TempDir(), "nope"), false)
	if err == nil {
		t.Fatal("expected error walking missing dir with ignoreErr=false, got nil")
	}
}

func TestWalkMissingDirIgnoreErr(t *testing.T) {
	t.Parallel()
	// With ignoreErr=true the walk function swallows the error.
	if _, _, err := Walk(filepath.Join(t.TempDir(), "nope"), true); err != nil {
		t.Fatalf("expected nil error with ignoreErr=true, got %v", err)
	}
}

func TestWalkListsDirsAndFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(sub, "f.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dirs, files, err := Walk(dir, false)
	if err != nil {
		t.Fatalf("Walk error = %v", err)
	}
	if !contains(dirs, dir) || !contains(dirs, sub) {
		t.Errorf("dirs = %v, want to include %q and %q", dirs, dir, sub)
	}
	if !contains(files, f) {
		t.Errorf("files = %v, want to include %q", files, f)
	}
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// path.go
// ---------------------------------------------------------------------------

func TestIsSamePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		src  string
		dest string
		want bool
	}{
		{"identical relative", "./a/b", "./a/b", true},
		{"relative vs equivalent", "a/b", "a/./b", true},
		{"different", "a/b", "a/c", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsSamePath(tt.src, tt.dest); got != tt.want {
				t.Errorf("IsSamePath(%q,%q) = %v, want %v", tt.src, tt.dest, got, tt.want)
			}
		})
	}
}

func TestTopDirName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want string
	}{
		{"foo/bar/baz", "foo"},
		{"single", "single"},
		{"/abs/path", ""},
	}
	for _, tt := range tests {
		if got := TopDirName(tt.path); got != tt.want {
			t.Errorf("TopDirName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// shell.go
// ---------------------------------------------------------------------------

func TestConcatenateMissingFile(t *testing.T) {
	t.Parallel()
	if _, err := Concatenate([]string{filepath.Join(t.TempDir(), "nope")}); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestConcatenateJoinsFilesWithoutTrailingNewline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// First file: two lines, the second without trailing newline.
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("a1\na2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("b1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Concatenate([]string{a, b})
	if err != nil {
		t.Fatalf("Concatenate error = %v", err)
	}
	joined := strings.Join(got, "|")
	// The last line of a ("a2") has no newline, so it is glued to b's first line.
	if !strings.Contains(joined, "a2b1") {
		t.Errorf("Concatenate = %v, expected a2 glued to b1", got)
	}
}

func TestGroupsUnknownUser(t *testing.T) {
	t.Parallel()
	// A user name that cannot exist.
	if _, err := Groups("this-user-should-not-exist-xyzzy"); err == nil {
		t.Fatal("expected error for unknown user, got nil")
	}
}

func TestHasOperandVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		cmdName string
		want    bool
	}{
		{"only short flags", []string{"cmd", "-a", "-b"}, "cmd", false},
		{"only long flags", []string{"cmd", "--all", "--verbose"}, "cmd", false},
		{"has operand file", []string{"cmd", "-a", "file.txt"}, "cmd", true},
		{"long flag treated as no operand", []string{"cmd", "--xyz"}, "cmd", false},
		{"three-char dash is operand", []string{"cmd", "-ab"}, "cmd", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasOperand(tt.args, tt.cmdName); got != tt.want {
				t.Errorf("HasOperand(%v,%q) = %v, want %v", tt.args, tt.cmdName, got, tt.want)
			}
			if got := HasNoOperand(tt.args, tt.cmdName); got == tt.want {
				t.Errorf("HasNoOperand should be the inverse of HasOperand")
			}
		})
	}
}

// withStdin temporarily replaces os.Stdin with a real file containing data,
// runs fn, then restores the original. It is used to exercise the stdin-reading
// helpers (Input, FromPIPE, HasPipeData) deterministically. Tests using it must
// not run in parallel because os.Stdin is process-global.
func withStdin(t *testing.T, data string, fn func()) {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stdin-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(data); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	orig := os.Stdin
	os.Stdin = f
	defer func() {
		os.Stdin = orig
		_ = f.Close()
	}()
	fn()
}

func TestInputReadsLine(t *testing.T) {
	withStdin(t, "answer\n", func() {
		got, ok := Input()
		if !ok {
			t.Fatalf("Input ok = false, want true")
		}
		if got != "answer" {
			t.Errorf("Input = %q, want %q", got, "answer")
		}
	})
}

func TestInputEOFReturnsFalse(t *testing.T) {
	withStdin(t, "", func() {
		got, ok := Input()
		if ok {
			t.Errorf("Input on empty stdin ok = true, want false (got %q)", got)
		}
	})
}

func TestHasPipeDataWithFileStdin(t *testing.T) {
	// A regular file is not a terminal, so HasPipeData must report true.
	withStdin(t, "piped\n", func() {
		if !HasPipeData() {
			t.Error("HasPipeData with file stdin = false, want true")
		}
	})
}

func TestFromPIPEReadsAll(t *testing.T) {
	withStdin(t, "line1\nline2\n", func() {
		got, err := FromPIPE()
		if err != nil {
			t.Fatalf("FromPIPE error = %v", err)
		}
		if got != "line1\nline2\n" {
			t.Errorf("FromPIPE = %q, want %q", got, "line1\nline2\n")
		}
	})
}

func TestWrapStringEdges(t *testing.T) {
	t.Parallel()
	if got := WrapString("abcdef", 0); got != "abcdef" {
		t.Errorf("WrapString with column<=0 should return src unchanged, got %q", got)
	}
	if got := WrapString("abcdef", 2); got != "ab\ncd\nef" {
		t.Errorf("WrapString = %q, want ab\\ncd\\nef", got)
	}
	if got := WrapString("abcde", 2); got != "ab\ncd\ne" {
		t.Errorf("WrapString = %q, want ab\\ncd\\ne", got)
	}
}
