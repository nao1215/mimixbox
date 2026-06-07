package df_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/df"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := df.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := df.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if got := c.Name(); got != "df" {
		t.Errorf("Name() = %q, want %q", got, "df")
	}
	if got := c.Synopsis(); got != "Report file system disk space usage" {
		t.Errorf("Synopsis() = %q", got)
	}
}

func TestComputeUsage(t *testing.T) {
	t.Parallel()
	// 1000 blocks of 1024 bytes = 1,024,000 total. 300 free blocks (307,200)
	// but only 250 available to non-root (256,000); the 50-block difference is
	// reserved. Used = total - free = 716,800 (GNU df counts reserved blocks as
	// used). Usable total = used+avail = 972,800; use% = ceil(73.68) = 74.
	s := df.StatfsResult{Bsize: 1024, Blocks: 1000, Bfree: 300, Bavail: 250}
	u := df.ComputeUsage(s)

	if u.Total() != 1024000 {
		t.Errorf("Total = %d, want 1024000", u.Total())
	}
	if u.Avail() != 256000 {
		t.Errorf("Avail = %d, want 256000", u.Avail())
	}
	if u.Used() != 716800 {
		t.Errorf("Used = %d, want 716800", u.Used())
	}
	if u.UsePct() != 74 {
		t.Errorf("UsePct = %d, want 74", u.UsePct())
	}
}

func TestComputeUsagePercentRoundsUp(t *testing.T) {
	t.Parallel()
	// used = blocks-bfree = 1, total(usable) = used+avail = 3 ->
	// 1*100/3 = 33.3 -> rounds up to 34.
	s := df.StatfsResult{Bsize: 1, Blocks: 3, Bfree: 2, Bavail: 2}
	u := df.ComputeUsage(s)
	if u.UsePct() != 34 {
		t.Errorf("UsePct = %d, want 34", u.UsePct())
	}
}

func TestComputeInodeUsage(t *testing.T) {
	t.Parallel()
	s := df.StatfsResult{Files: 1000, Ffree: 600}
	u := df.ComputeInodeUsage(s)
	if u.Files() != 1000 {
		t.Errorf("Files = %d, want 1000", u.Files())
	}
	if u.IUsed() != 400 {
		t.Errorf("IUsed = %d, want 400", u.IUsed())
	}
	if u.IFree() != 600 {
		t.Errorf("IFree = %d, want 600", u.IFree())
	}
	if u.IUsePct() != 40 {
		t.Errorf("IUsePct = %d, want 40", u.IUsePct())
	}
}

func TestHumanReadable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   uint64
		want string
	}{
		{0, "0"},
		{512, "512"},
		{1023, "1023"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
		{1073741824, "1.0G"},
		{1099511627776, "1.0T"},
	}
	for _, tt := range tests {
		if got := df.HumanReadable(tt.in); got != tt.want {
			t.Errorf("HumanReadable(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatSize(t *testing.T) {
	t.Parallel()
	// 1K blocks: 1,024,000 bytes / 1024 = 1000.
	if got := df.FormatSize(1024000, false); got != "1000" {
		t.Errorf("FormatSize(1024000,false) = %q, want %q", got, "1000")
	}
	// Rounds up partial blocks.
	if got := df.FormatSize(1, false); got != "1" {
		t.Errorf("FormatSize(1,false) = %q, want %q", got, "1")
	}
	if got := df.FormatSize(1048576, true); got != "1.0M" {
		t.Errorf("FormatSize(1048576,true) = %q, want %q", got, "1.0M")
	}
}

// TestRunWithFake exercises the full Run path against an injected fake statfs so
// the numeric output is deterministic.
func TestRunWithFake(t *testing.T) {
	restore := df.SetStatfs(func(path string) (df.StatfsResult, error) {
		return df.StatfsResult{Bsize: 1024, Blocks: 1000, Bfree: 250, Bavail: 250, Files: 100, Ffree: 60}, nil
	})
	defer restore()

	out, errOut, err := run(t, "/data")
	if err != nil {
		t.Fatalf("Run error: %v (stderr=%q)", err, errOut)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected header + 1 row, got %d lines: %q", len(lines), out)
	}
	if !strings.HasPrefix(lines[0], "Filesystem") {
		t.Errorf("header = %q, want it to start with Filesystem", lines[0])
	}
	if !strings.Contains(lines[0], "1K-blocks") {
		t.Errorf("header = %q, want 1K-blocks column", lines[0])
	}
	fields := strings.Fields(lines[1])
	// /data 1000 750 250 75% /data
	want := []string{"/data", "1000", "750", "250", "75%", "/data"}
	if len(fields) != len(want) {
		t.Fatalf("row fields = %v, want %v", fields, want)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Errorf("row field %d = %q, want %q (row=%q)", i, fields[i], want[i], lines[1])
		}
	}
}

func TestRunHumanReadable(t *testing.T) {
	restore := df.SetStatfs(func(path string) (df.StatfsResult, error) {
		// 1 MiB total, 0 available.
		return df.StatfsResult{Bsize: 1024, Blocks: 1024, Bavail: 0}, nil
	})
	defer restore()

	out, _, err := run(t, "-h", "/data")
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if !strings.Contains(lines[0], "Size") {
		t.Errorf("human header = %q, want Size column", lines[0])
	}
	fields := strings.Fields(lines[1])
	if fields[1] != "1.0M" {
		t.Errorf("human size = %q, want 1.0M", fields[1])
	}
}

func TestRunInodes(t *testing.T) {
	restore := df.SetStatfs(func(path string) (df.StatfsResult, error) {
		return df.StatfsResult{Files: 100, Ffree: 60}, nil
	})
	defer restore()

	out, _, err := run(t, "-i", "/data")
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if !strings.Contains(lines[0], "Inodes") {
		t.Errorf("inode header = %q, want Inodes column", lines[0])
	}
	fields := strings.Fields(lines[1])
	// /data 100 40 60 40% /data
	want := []string{"/data", "100", "40", "60", "40%", "/data"}
	for i := range want {
		if fields[i] != want[i] {
			t.Errorf("inode row field %d = %q, want %q", i, fields[i], want[i])
		}
	}
}

// TestRunRealEndToEnd checks that "df ." against the real filesystem succeeds
// and prints a header. Numeric values are not asserted here.
func TestRunRealEndToEnd(t *testing.T) {
	out, errOut, err := run(t, ".")
	if err != nil {
		t.Fatalf("Run(.) error: %v (stderr=%q)", err, errOut)
	}
	if !strings.HasPrefix(out, "Filesystem") {
		t.Errorf("output should start with header, got %q", out)
	}
}

func TestRunNoOperandDefaultsToCwd(t *testing.T) {
	restore := df.SetStatfs(func(path string) (df.StatfsResult, error) {
		if path != "." {
			t.Errorf("expected default operand %q, got %q", ".", path)
		}
		return df.StatfsResult{Bsize: 1024, Blocks: 10, Bavail: 5}, nil
	})
	defer restore()

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !strings.HasPrefix(out, "Filesystem") {
		t.Errorf("output should start with header, got %q", out)
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("Run(--help) error: %v", err)
	}
	if !strings.Contains(out, "Usage: df") {
		t.Errorf("help output missing usage, got %q", out)
	}
}
