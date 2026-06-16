package tail

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func newIO(in string) (command.IO, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	return command.IO{In: strings.NewReader(in), Out: out, Err: errBuf}, out, errBuf
}

// TestNewFollowTargets covers the open/seek/retry branches without entering the
// real-time follow loop.
func TestNewFollowTargets(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	existing := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(existing, []byte("hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing")

	// Without retry the missing path is dropped, the existing one is opened
	// positioned at EOF.
	targets := newFollowTargets([]string{existing, missing}, false)
	defer closeAll(targets)
	if len(targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(targets))
	}
	if targets[0].offset != int64(len("hello\n")) {
		t.Errorf("offset = %d, want %d", targets[0].offset, len("hello\n"))
	}

	// With retry the missing path is kept as a pending (file == nil) target.
	retryTargets := newFollowTargets([]string{missing}, true)
	defer closeAll(retryTargets)
	if len(retryTargets) != 1 || retryTargets[0].file != nil {
		t.Fatalf("retry targets = %+v, want one pending target", retryTargets)
	}
}

// TestPollEmitsAppendedData appends to a followed file and polls once, checking
// the new bytes are emitted and the offset advances.
func TestPollEmitsAppendedData(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	targets := newFollowTargets([]string{path}, false)
	defer closeAll(targets)

	if err := os.WriteFile(path, []byte("a\nb\nc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	io, out, _ := newIO("")
	last := ""
	targets[0].poll(io, false, false, &last)
	if out.String() != "b\nc\n" {
		t.Errorf("poll output = %q, want %q", out.String(), "b\nc\n")
	}
	if targets[0].offset != int64(len("a\nb\nc\n")) {
		t.Errorf("offset = %d, want %d", targets[0].offset, len("a\nb\nc\n"))
	}
}

// TestPollDetectsTruncation verifies a shrunk file restarts from offset 0 and
// reports the truncation on stderr.
func TestPollDetectsTruncation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("longcontent\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	targets := newFollowTargets([]string{path}, false)
	defer closeAll(targets)

	if err := os.WriteFile(path, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	io, out, errBuf := newIO("")
	last := ""
	targets[0].poll(io, false, false, &last)
	if !strings.Contains(errBuf.String(), "file truncated") {
		t.Errorf("stderr = %q, want truncation notice", errBuf.String())
	}
	if !strings.Contains(out.String(), "x\n") {
		t.Errorf("output = %q, want restarted content", out.String())
	}
}

// TestMaybeReopenOnReplacement covers the -F rotation path: the followed file is
// renamed away and recreated, and maybeReopen switches to the new descriptor.
func TestMaybeReopenOnReplacement(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("first\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	targets := newFollowTargets([]string{path}, true)
	defer closeAll(targets)

	if err := os.Rename(path, path+".old"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("second\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	io, _, errBuf := newIO("")
	targets[0].maybeReopen(io)
	if targets[0].offset != 0 {
		t.Errorf("offset after reopen = %d, want 0", targets[0].offset)
	}
	if !strings.Contains(errBuf.String(), "following new file") {
		t.Errorf("stderr = %q, want reopen notice", errBuf.String())
	}
}

// TestMaybeReopenWhenMissing covers the branch where the path disappears: the
// held descriptor is released and the target becomes pending.
func TestMaybeReopenWhenMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	targets := newFollowTargets([]string{path}, true)
	defer closeAll(targets)

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	io, _, _ := newIO("")
	targets[0].maybeReopen(io)
	if targets[0].file != nil {
		t.Error("file should be nil after the path disappears")
	}

	// A pending target whose path reappears should be picked up.
	if err := os.WriteFile(path, []byte("back\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	io2, _, errBuf := newIO("")
	targets[0].maybeReopen(io2)
	if targets[0].file == nil {
		t.Error("file should be reopened after the path reappears")
	}
	if !strings.Contains(errBuf.String(), "has appeared") {
		t.Errorf("stderr = %q, want appearance notice", errBuf.String())
	}
}

// TestFollowablePaths drops the stdin pseudo-file.
func TestFollowablePaths(t *testing.T) {
	t.Parallel()
	got := followablePaths([]string{"-", "a.txt", "-", "b.txt"})
	if len(got) != 2 || got[0] != "a.txt" || got[1] != "b.txt" {
		t.Errorf("followablePaths = %v, want [a.txt b.txt]", got)
	}
}

// TestWriteHeader covers both the first and subsequent header forms, including
// the standard-input label.
func TestWriteHeader(t *testing.T) {
	t.Parallel()
	var first bytes.Buffer
	writeHeader(&first, "a.txt", true)
	if first.String() != "==> a.txt <==\n" {
		t.Errorf("first header = %q", first.String())
	}
	var later bytes.Buffer
	writeHeader(&later, "-", false)
	if later.String() != "\n==> standard input <==\n" {
		t.Errorf("later header = %q", later.String())
	}
}

// TestEmitWritesHeaderOnFileSwitch checks emit prints a banner when output
// switches between files.
func TestEmitWritesHeaderOnFileSwitch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("data\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	tgt := &followTarget{path: path, file: f, offset: 0}

	io, out, _ := newIO("")
	last := ""
	tgt.emit(io, int64(len("data\n")), true, &last)
	got := out.String()
	if !strings.Contains(got, "==> "+path+" <==") || !strings.Contains(got, "data\n") {
		t.Errorf("emit output = %q, want header and data", got)
	}
	if last != path {
		t.Errorf("last = %q, want %q", last, path)
	}
}

func TestSynopsis(t *testing.T) {
	t.Parallel()
	if New().Synopsis() == "" {
		t.Error("Synopsis() = empty")
	}
}
