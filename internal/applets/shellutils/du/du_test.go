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
