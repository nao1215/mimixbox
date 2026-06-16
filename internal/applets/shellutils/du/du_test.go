package du_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/du"
	"github.com/nao1215/mimixbox/internal/command"
)

// runDu runs the du applet against the given args and returns stdout, stderr,
// and the error.
func runDu(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := du.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// buildTree creates a known directory tree:
//
//	root/a.txt      1000 bytes -> 1 block
//	root/b.txt      2000 bytes -> 2 blocks
//	root/sub/c.txt  3000 bytes -> 3 blocks
//
// Total apparent size: 6000 bytes -> 6 blocks (ceil(6000/1024) = 6).
func buildTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o750); err != nil {
		t.Fatal(err)
	}
	write := func(path string, n int) {
		if err := os.WriteFile(path, bytes.Repeat([]byte("x"), n), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "a.txt"), 1000)
	write(filepath.Join(root, "b.txt"), 2000)
	write(filepath.Join(sub, "c.txt"), 3000)
	return root
}

func TestSummarizePrintsSingleTotal(t *testing.T) {
	t.Parallel()
	root := buildTree(t)

	out, errOut, err := runDu(t, "-s", root)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	lines := nonEmptyLines(out)
	if len(lines) != 1 {
		t.Fatalf("-s should print exactly one line, got %d: %q", len(lines), out)
	}
	size, path := splitLine(t, lines[0])
	if size != "6" {
		t.Errorf("size = %q, want %q (6 blocks)", size, "6")
	}
	if path != root {
		t.Errorf("path = %q, want %q", path, root)
	}
}

func TestBytesPrintsByteSizes(t *testing.T) {
	t.Parallel()
	root := buildTree(t)

	out, errOut, err := runDu(t, "-s", "-b", root)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	lines := nonEmptyLines(out)
	if len(lines) != 1 {
		t.Fatalf("expected one line, got %d: %q", len(lines), out)
	}
	size, _ := splitLine(t, lines[0])
	// Apparent bytes: 1000 + 2000 + 3000 = 6000.
	if size != "6000" {
		t.Errorf("size = %q, want %q bytes", size, "6000")
	}
}

func TestAllListsIndividualFiles(t *testing.T) {
	t.Parallel()
	root := buildTree(t)

	out, errOut, err := runDu(t, "-a", root)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}

	want := map[string]string{
		filepath.Join(root, "a.txt"):        "1",
		filepath.Join(root, "b.txt"):        "2",
		filepath.Join(root, "sub", "c.txt"): "3",
		filepath.Join(root, "sub"):          "3",
		root:                                "6",
	}
	got := map[string]string{}
	for _, l := range nonEmptyLines(out) {
		size, path := splitLine(t, l)
		got[path] = size
	}
	for path, size := range want {
		if got[path] != size {
			t.Errorf("path %q size = %q, want %q (full output:\n%s)", path, got[path], size, out)
		}
	}
	// -a must include the individual files, which the default mode omits.
	if _, ok := got[filepath.Join(root, "a.txt")]; !ok {
		t.Errorf("-a did not list individual file a.txt: %q", out)
	}
}

func TestDefaultListsOnlyDirectories(t *testing.T) {
	t.Parallel()
	root := buildTree(t)

	out, errOut, err := runDu(t, root)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	for _, l := range nonEmptyLines(out) {
		_, path := splitLine(t, l)
		if strings.HasSuffix(path, ".txt") {
			t.Errorf("default mode listed a file: %q", l)
		}
	}
}

func TestTotalPrintsGrandTotal(t *testing.T) {
	t.Parallel()
	root := buildTree(t)

	out, errOut, err := runDu(t, "-s", "-c", root)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	lines := nonEmptyLines(out)
	last := lines[len(lines)-1]
	size, path := splitLine(t, last)
	if path != "total" {
		t.Fatalf("last line path = %q, want %q", path, "total")
	}
	if size != "6" {
		t.Errorf("grand total = %q, want %q", size, "6")
	}
}

func TestHumanReadableFlag(t *testing.T) {
	t.Parallel()
	root := buildTree(t)

	out, errOut, err := runDu(t, "-s", "-h", root)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	size, _ := splitLine(t, nonEmptyLines(out)[0])
	// 6000 bytes -> 5.9K (6000/1024 = 5.86).
	if size != "5.9K" {
		t.Errorf("human size = %q, want %q", size, "5.9K")
	}
}

// TestHumanReadableFormatter exercises the human-readable formatting directly
// through the -h output for several deterministic sizes.
func TestHumanReadableFormatter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		size int
		want string
	}{
		{1536, "1.5K"},        // 1536 bytes -> 1.5K
		{1024, "1.0K"},        // exactly one block -> 1.0K
		{1024 * 1024, "1.0M"}, // one mebibyte
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			f := filepath.Join(dir, "f.bin")
			if err := os.WriteFile(f, bytes.Repeat([]byte("y"), tt.size), 0o600); err != nil {
				t.Fatal(err)
			}
			out, _, err := runDu(t, "-h", f)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			size, _ := splitLine(t, nonEmptyLines(out)[0])
			if size != tt.want {
				t.Errorf("size for %d bytes = %q, want %q", tt.size, size, tt.want)
			}
		})
	}
}

// TestSynopsisAndName covers the metadata accessors.
func TestSynopsisAndName(t *testing.T) {
	t.Parallel()
	c := du.New()
	if c.Name() != "du" {
		t.Errorf("Name() = %q, want du", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestEmptyFileZeroBlocks covers the blocks(bytes<=0) branch: a zero-byte file
// occupies zero 1K blocks.
func TestEmptyFileZeroBlocks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(f, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := runDu(t, f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	size, _ := splitLine(t, nonEmptyLines(out)[0])
	if size != "0" {
		t.Errorf("size = %q, want %q for empty file", size, "0")
	}
}

// TestHumanReadableNoDecimalAboveTen covers the %.0f branch of humanReadable,
// taken when the scaled value is >= 10 (here ~15K).
func TestHumanReadableNoDecimalAboveTen(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "big.bin")
	if err := os.WriteFile(f, bytes.Repeat([]byte("z"), 15*1024), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := runDu(t, "-h", f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	size, _ := splitLine(t, nonEmptyLines(out)[0])
	if size != "15K" {
		t.Errorf("size = %q, want %q (no decimal place at/above 10)", size, "15K")
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := runDu(t, "/no/such/path")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
	if !strings.Contains(errOut, "du: /no/such/path:") {
		t.Errorf("stderr = %q, want du error prefix", errOut)
	}
}

func TestHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := runDu(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: du") {
		t.Errorf("--help out = %q", out)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q:\n%s", want, out)
		}
	}

	out, _, err = runDu(t, "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "du (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}

// buildDeepTree creates:
//
//	root/a.txt          (1000 bytes)
//	root/sub/b.txt      (2000 bytes)
//	root/sub/deep/c.txt (3000 bytes)
//
// so depth 0 = root, depth 1 = root/sub, depth 2 = root/sub/deep.
func buildDeepTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	deep := filepath.Join(root, "sub", "deep")
	if err := os.MkdirAll(deep, 0o750); err != nil {
		t.Fatal(err)
	}
	write := func(p string, n int) {
		if err := os.WriteFile(p, bytes.Repeat([]byte("x"), n), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "a.txt"), 1000)
	write(filepath.Join(root, "sub", "b.txt"), 2000)
	write(filepath.Join(deep, "c.txt"), 3000)
	return root
}

func TestMaxDepthOmitsDeeperDirectories(t *testing.T) {
	t.Parallel()
	root := buildDeepTree(t)

	out, errOut, err := runDu(t, "--max-depth=1", root)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	got := map[string]bool{}
	for _, l := range nonEmptyLines(out) {
		_, p := splitLine(t, l)
		got[p] = true
	}
	// depth 0 (root) and depth 1 (root/sub) must be present.
	if !got[root] {
		t.Errorf("--max-depth=1 omitted the operand root: %q", out)
	}
	if !got[filepath.Join(root, "sub")] {
		t.Errorf("--max-depth=1 omitted depth-1 dir sub: %q", out)
	}
	// depth 2 (root/sub/deep) must be omitted.
	if got[filepath.Join(root, "sub", "deep")] {
		t.Errorf("--max-depth=1 should omit depth-2 dir deep: %q", out)
	}
}

func TestMaxDepthZeroPrintsOnlyRoot(t *testing.T) {
	t.Parallel()
	root := buildDeepTree(t)

	out, errOut, err := runDu(t, "--max-depth=0", root)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	lines := nonEmptyLines(out)
	if len(lines) != 1 {
		t.Fatalf("--max-depth=0 should print one line, got %d: %q", len(lines), out)
	}
	_, p := splitLine(t, lines[0])
	if p != root {
		t.Errorf("--max-depth=0 line = %q, want root %q", p, root)
	}
}

func TestExcludeSkipsMatchingEntries(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write := func(name string, n int) {
		if err := os.WriteFile(filepath.Join(root, name), bytes.Repeat([]byte("x"), n), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("keep.txt", 1024)
	write("drop.tmp", 4096)

	out, errOut, err := runDu(t, "-a", "--exclude=*.tmp", root)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	for _, l := range nonEmptyLines(out) {
		_, p := splitLine(t, l)
		if strings.HasSuffix(p, "drop.tmp") {
			t.Errorf("--exclude=*.tmp did not skip drop.tmp: %q", l)
		}
	}
	// The excluded file's size must not be counted toward the root total.
	rootBlocks := ""
	for _, l := range nonEmptyLines(out) {
		size, p := splitLine(t, l)
		if p == root {
			rootBlocks = size
		}
	}
	// Only keep.txt (1024 bytes -> 1 block) should count.
	if rootBlocks != "1" {
		t.Errorf("root total = %q blocks, want %q (excluded file must not count)", rootBlocks, "1")
	}
}

func TestExcludeDirectoryPrunesSubtree(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	skip := filepath.Join(root, "skipme")
	if err := os.Mkdir(skip, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skip, "big.bin"), bytes.Repeat([]byte("x"), 8192), 0o600); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := runDu(t, "--exclude=skipme", root)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	for _, l := range nonEmptyLines(out) {
		_, p := splitLine(t, l)
		if strings.Contains(p, "skipme") {
			t.Errorf("--exclude=skipme did not prune the directory: %q", l)
		}
	}
}

func TestApparentSizeDiffersFromBlockRounding(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// A single 3000-byte file: the default reports 3 1K blocks (ceil(3000/1024)),
	// while --apparent-size reports the exact byte count 3000.
	if err := os.WriteFile(filepath.Join(root, "f.bin"), bytes.Repeat([]byte("a"), 3000), 0o600); err != nil {
		t.Fatal(err)
	}

	defOut, _, err := runDu(t, "-s", root)
	if err != nil {
		t.Fatalf("default Run error = %v", err)
	}
	appOut, _, err := runDu(t, "-s", "--apparent-size", root)
	if err != nil {
		t.Fatalf("--apparent-size Run error = %v", err)
	}

	defSize, _ := splitLine(t, nonEmptyLines(defOut)[0])
	appSize, _ := splitLine(t, nonEmptyLines(appOut)[0])

	// Default: block count, ceil(3000/1024) = 3.
	if defSize != "3" {
		t.Errorf("default blocks = %q, want %q", defSize, "3")
	}
	// --apparent-size: exact byte count.
	if appSize != "3000" {
		t.Errorf("--apparent-size = %q, want %q (exact bytes)", appSize, "3000")
	}
	if defSize == appSize {
		t.Errorf("--apparent-size (%q) should differ from default block count (%q)", appSize, defSize)
	}
}

func TestApparentSizeBytes(t *testing.T) {
	t.Parallel()
	root := buildTree(t)

	out, _, err := runDu(t, "-s", "--apparent-size", "-b", root)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	size, _ := splitLine(t, nonEmptyLines(out)[0])
	if size != "6000" {
		t.Errorf("apparent bytes = %q, want %q", size, "6000")
	}
}

func TestOneFileSystemSameDevice(t *testing.T) {
	t.Parallel()
	// Everything in a temp tree lives on one filesystem, so -x must produce the
	// same output as a plain run: no boundary is crossed.
	root := buildTree(t)

	plain, _, err := runDu(t, root)
	if err != nil {
		t.Fatalf("plain Run error = %v", err)
	}
	withX, errOut, err := runDu(t, "-x", root)
	if err != nil {
		t.Fatalf("-x Run error = %v (stderr=%q)", err, errOut)
	}
	if plain != withX {
		t.Errorf("-x changed output on a single filesystem:\nplain=%q\n-x   =%q", plain, withX)
	}
}

func TestOneFileSystemLongFlag(t *testing.T) {
	t.Parallel()
	root := buildTree(t)
	out, errOut, err := runDu(t, "--one-file-system", "-s", root)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	size, _ := splitLine(t, nonEmptyLines(out)[0])
	if size != "6" {
		t.Errorf("--one-file-system total = %q, want %q", size, "6")
	}
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

func splitLine(t *testing.T, line string) (size, path string) {
	t.Helper()
	parts := strings.SplitN(line, "\t", 2)
	if len(parts) != 2 {
		t.Fatalf("line %q is not in SIZE\\tPATH form", line)
	}
	return parts[0], parts[1]
}
