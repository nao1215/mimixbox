package df_test

import (
	"bytes"
	"context"
	"errors"
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

// TestPercentZeroTotal covers the total==0 guard in percent via an empty
// filesystem (no usable blocks), which must report 0% rather than divide by 0.
func TestPercentZeroTotal(t *testing.T) {
	s := df.StatfsResult{Bsize: 1024, Blocks: 0, Bfree: 0, Bavail: 0}
	u := df.ComputeUsage(s)
	if u.UsePct() != 0 {
		t.Errorf("UsePct = %d, want 0 for empty filesystem", u.UsePct())
	}
}

// TestRunStatfsError exercises the error path of Run (the keep() helper) by
// injecting a statfs that fails, and confirms a non-zero exit while the header
// is still printed.
func TestRunStatfsError(t *testing.T) {
	restore := df.SetStatfs(func(path string) (df.StatfsResult, error) {
		return df.StatfsResult{}, errors.New("statfs boom")
	})
	defer restore()

	out, errOut, err := run(t, "/data")
	if err == nil {
		t.Fatal("expected non-nil error when statfs fails")
	}
	if !strings.HasPrefix(out, "Filesystem") {
		t.Errorf("header should still be printed, got %q", out)
	}
	if !strings.Contains(errOut, "df:") {
		t.Errorf("stderr = %q, want df: prefix", errOut)
	}
}

// TestRunStatfsErrorThenSuccess confirms keep() preserves the first error while
// later operands are still processed.
func TestRunStatfsErrorThenSuccess(t *testing.T) {
	restore := df.SetStatfs(func(path string) (df.StatfsResult, error) {
		if path == "/bad" {
			return df.StatfsResult{}, errors.New("nope")
		}
		return df.StatfsResult{Bsize: 1024, Blocks: 10, Bavail: 5}, nil
	})
	defer restore()

	out, _, err := run(t, "/bad", "/good")
	if err == nil {
		t.Fatal("expected non-nil error from the failing operand")
	}
	if !strings.Contains(out, "/good") {
		t.Errorf("the good operand should still be printed, got %q", out)
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

// ---- GNU issue #754 additions ----

func TestParseSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"1024", 1024, true},
		{"K", 1024, true},
		{"1K", 1024, true},
		{"1M", 1024 * 1024, true},
		{"1G", 1024 * 1024 * 1024, true},
		{"512", 512, true},
		{"1MB", 1000 * 1000, true},
		{"", 0, false},
		{"0", 0, false},
		{"1Z", 0, false},
	}
	for _, tt := range tests {
		got, err := df.ParseSize(tt.in)
		if tt.ok && (err != nil || got != tt.want) {
			t.Errorf("ParseSize(%q) = %d, %v; want %d", tt.in, got, err, tt.want)
		}
		if !tt.ok && err == nil {
			t.Errorf("ParseSize(%q) expected error", tt.in)
		}
	}
}

func TestScaleSizeBlockSize(t *testing.T) {
	t.Parallel()
	// 1 MiB with a 1M block-size = 1 block.
	if got := df.ScaleSize(1024*1024, false, 1024*1024); got != "1" {
		t.Errorf("ScaleSize(1M, bs=1M) = %q, want 1", got)
	}
	// 1 MiB in default 1K blocks = 1024.
	if got := df.ScaleSize(1024*1024, false, 0); got != "1024" {
		t.Errorf("ScaleSize(1M, bs=0) = %q, want 1024", got)
	}
	// Rounds up partial blocks.
	if got := df.ScaleSize(1, false, 1024); got != "1" {
		t.Errorf("ScaleSize(1, bs=1K) = %q, want 1", got)
	}
	// Human-readable ignores block-size.
	if got := df.ScaleSize(1024*1024, true, 1024); got != "1.0M" {
		t.Errorf("ScaleSize(1M, human) = %q, want 1.0M", got)
	}
}

func TestParseOutput(t *testing.T) {
	t.Parallel()
	cols, err := df.ParseOutput("source,fstype,size,used,avail,pcent,target")
	if err != nil {
		t.Fatalf("ParseOutput error: %v", err)
	}
	want := []string{"source", "fstype", "size", "used", "avail", "pcent", "target"}
	if len(cols) != len(want) {
		t.Fatalf("cols = %v, want %v", cols, want)
	}
	for i := range want {
		if cols[i] != want[i] {
			t.Errorf("col %d = %q, want %q", i, cols[i], want[i])
		}
	}
	// Order is preserved (reversed).
	rev, _ := df.ParseOutput("target,pcent,size")
	if rev[0] != "target" || rev[2] != "size" {
		t.Errorf("ParseOutput did not preserve order: %v", rev)
	}
	// Unknown field is an error.
	if _, err := df.ParseOutput("bogus"); err == nil {
		t.Error("ParseOutput(bogus) expected error")
	}
}

func TestFilterByType(t *testing.T) {
	t.Parallel()
	entries := []df.FsEntry{
		df.NewFsEntry("/dev/sda1", "ext4", "/", df.StatfsResult{Bsize: 1024, Blocks: 10}),
		df.NewFsEntry("tmpfs", "tmpfs", "/run", df.StatfsResult{Bsize: 1024, Blocks: 5}),
		df.NewFsEntry("/dev/sdb1", "ext4", "/home", df.StatfsResult{Bsize: 1024, Blocks: 20}),
	}
	got := df.FilterByType(entries, []string{"ext4"})
	want := []string{"/", "/home"}
	if len(got) != len(want) {
		t.Fatalf("filtered targets = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("target %d = %q, want %q", i, got[i], want[i])
		}
	}
	// Empty filter keeps everything.
	if all := df.FilterByType(entries, nil); len(all) != 3 {
		t.Errorf("no filter kept %d, want 3", len(all))
	}
	// Repeatable filter matching multiple types.
	if both := df.FilterByType(entries, []string{"ext4", "tmpfs"}); len(both) != 3 {
		t.Errorf("ext4+tmpfs kept %d, want 3", len(both))
	}
}

func TestUnescapeMount(t *testing.T) {
	t.Parallel()
	if got := df.UnescapeMount(`/mnt/my\040disk`); got != "/mnt/my disk" {
		t.Errorf("UnescapeMount = %q, want %q", got, "/mnt/my disk")
	}
	if got := df.UnescapeMount("/plain/path"); got != "/plain/path" {
		t.Errorf("UnescapeMount = %q, want %q", got, "/plain/path")
	}
}

func TestFsTypeName(t *testing.T) {
	t.Parallel()
	if got := df.FsTypeName(0xef53); got != "ext" {
		t.Errorf("FsTypeName(ext magic) = %q, want ext", got)
	}
	if got := df.FsTypeName(0x12345); got != "0x12345" {
		t.Errorf("FsTypeName(unknown) = %q, want hex fallback", got)
	}
}

// withFakes installs deterministic statfs+mount fakes and returns a restore.
func withFakes(t *testing.T, mounts []df.MountEntry, stats map[string]df.StatfsResult) func() {
	t.Helper()
	r1 := df.SetReadMounts(func() ([]df.MountEntry, error) { return mounts, nil })
	r2 := df.SetStatfs(func(path string) (df.StatfsResult, error) {
		if s, ok := stats[path]; ok {
			return s, nil
		}
		return df.StatfsResult{Bsize: 1024, Blocks: 10, Bavail: 5}, nil
	})
	return func() { r2(); r1() }
}

func TestRunOutputColumnSelection(t *testing.T) {
	mounts := []df.MountEntry{
		df.NewMountEntry("/dev/sda1", "/", "ext4"),
		df.NewMountEntry("tmpfs", "/run", "tmpfs"),
	}
	stats := map[string]df.StatfsResult{
		"/":    {Bsize: 1024, Blocks: 1000, Bfree: 250, Bavail: 250},
		"/run": {Bsize: 1024, Blocks: 100, Bfree: 100, Bavail: 100},
	}
	defer withFakes(t, mounts, stats)()

	out, errOut, err := run(t, "--output=source,fstype,size,used,avail,pcent,target")
	if err != nil {
		t.Fatalf("Run error: %v (stderr=%q)", err, errOut)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	header := strings.Fields(lines[0])
	wantHeader := []string{"Filesystem", "Type", "Size", "Used", "Avail", "Use%", "Mounted", "on"}
	// "Mounted on" splits into two fields; compare the meaningful prefix.
	for i, h := range []string{"Filesystem", "Type", "Size", "Used", "Avail", "Use%"} {
		if header[i] != h {
			t.Errorf("header[%d] = %q, want %q (header=%q)", i, header[i], h, lines[0])
		}
	}
	_ = wantHeader
	// Find the / row.
	var rootFields []string
	for _, l := range lines[1:] {
		f := strings.Fields(l)
		if len(f) > 0 && f[0] == "/dev/sda1" {
			rootFields = f
		}
	}
	if rootFields == nil {
		t.Fatalf("no /dev/sda1 row in:\n%s", out)
	}
	// source=/dev/sda1 fstype=ext4 ... target=/
	if rootFields[1] != "ext4" {
		t.Errorf("fstype = %q, want ext4", rootFields[1])
	}
	if rootFields[len(rootFields)-1] != "/" {
		t.Errorf("target = %q, want /", rootFields[len(rootFields)-1])
	}
}

func TestRunOutputColumnOrder(t *testing.T) {
	mounts := []df.MountEntry{df.NewMountEntry("/dev/sda1", "/", "ext4")}
	stats := map[string]df.StatfsResult{"/": {Bsize: 1024, Blocks: 1000, Bavail: 500}}
	defer withFakes(t, mounts, stats)()

	out, _, err := run(t, "--output=target,fstype,source")
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	header := strings.Fields(lines[0])
	if header[0] != "Mounted" || header[2] != "Type" || header[3] != "Filesystem" {
		t.Errorf("reordered header = %q", lines[0])
	}
	row := strings.Fields(lines[1])
	if row[0] != "/" || row[1] != "ext4" || row[2] != "/dev/sda1" {
		t.Errorf("reordered row = %q", lines[1])
	}
}

func TestRunTypeFilter(t *testing.T) {
	mounts := []df.MountEntry{
		df.NewMountEntry("/dev/sda1", "/", "ext4"),
		df.NewMountEntry("tmpfs", "/run", "tmpfs"),
		df.NewMountEntry("/dev/sdb1", "/home", "ext4"),
	}
	stats := map[string]df.StatfsResult{
		"/":     {Bsize: 1024, Blocks: 1000, Bavail: 500},
		"/run":  {Bsize: 1024, Blocks: 100, Bavail: 100},
		"/home": {Bsize: 1024, Blocks: 2000, Bavail: 1000},
	}
	defer withFakes(t, mounts, stats)()

	out, _, err := run(t, "-t", "ext4")
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if strings.Contains(out, "tmpfs") || strings.Contains(out, "/run") {
		t.Errorf("tmpfs should be filtered out:\n%s", out)
	}
	if !strings.Contains(out, "/dev/sda1") || !strings.Contains(out, "/dev/sdb1") {
		t.Errorf("ext4 rows missing:\n%s", out)
	}
}

func TestRunTotalRow(t *testing.T) {
	mounts := []df.MountEntry{
		df.NewMountEntry("/dev/sda1", "/", "ext4"),
		df.NewMountEntry("/dev/sdb1", "/home", "ext4"),
	}
	stats := map[string]df.StatfsResult{
		// total=1000K used=(1000-500)=500K avail=500K
		"/": {Bsize: 1024, Blocks: 1000, Bfree: 500, Bavail: 500},
		// total=2000K used=(2000-1000)=1000K avail=1000K
		"/home": {Bsize: 1024, Blocks: 2000, Bfree: 1000, Bavail: 1000},
	}
	defer withFakes(t, mounts, stats)()

	out, _, err := run(t, "--total")
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	last := strings.Fields(lines[len(lines)-1])
	if last[0] != "total" {
		t.Fatalf("last row = %q, want a total row", lines[len(lines)-1])
	}
	// total 1K-blocks = 1000+2000 = 3000; used = 500+1000 = 1500; avail = 1500.
	if last[1] != "3000" || last[2] != "1500" || last[3] != "1500" {
		t.Errorf("total row = %v, want size=3000 used=1500 avail=1500", last)
	}
}

func TestRunTotalRowWithOutput(t *testing.T) {
	mounts := []df.MountEntry{
		df.NewMountEntry("/dev/sda1", "/", "ext4"),
		df.NewMountEntry("/dev/sdb1", "/home", "ext4"),
	}
	stats := map[string]df.StatfsResult{
		"/":     {Bsize: 1024, Blocks: 1000, Bfree: 500, Bavail: 500},
		"/home": {Bsize: 1024, Blocks: 2000, Bfree: 1000, Bavail: 1000},
	}
	defer withFakes(t, mounts, stats)()

	out, _, err := run(t, "--total", "--output=source,size,used,avail,target")
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	last := strings.Fields(lines[len(lines)-1])
	if last[0] != "total" {
		t.Fatalf("output total row = %q", lines[len(lines)-1])
	}
	if last[1] != "3000" || last[2] != "1500" || last[3] != "1500" {
		t.Errorf("output total row = %v, want 3000/1500/1500", last)
	}
}

func TestRunBlockSizeScaling(t *testing.T) {
	mounts := []df.MountEntry{df.NewMountEntry("/dev/sda1", "/", "ext4")}
	stats := map[string]df.StatfsResult{
		// total = 4 MiB, avail = 0.
		"/": {Bsize: 1024, Blocks: 4096, Bavail: 0},
	}
	defer withFakes(t, mounts, stats)()

	out, _, err := run(t, "--block-size=1M", "--output=size,target")
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	row := strings.Fields(lines[1])
	// 4 MiB / 1M = 4.
	if row[0] != "4" {
		t.Errorf("block-size scaled size = %q, want 4 (row=%q)", row[0], lines[1])
	}
}

func TestRunAllInclusion(t *testing.T) {
	mounts := []df.MountEntry{
		df.NewMountEntry("/dev/sda1", "/", "ext4"),
		df.NewMountEntry("proc", "/proc", "proc"), // zero-size pseudo fs
	}
	stats := map[string]df.StatfsResult{
		"/":     {Bsize: 1024, Blocks: 1000, Bavail: 500},
		"/proc": {Bsize: 1024, Blocks: 0, Bavail: 0}, // hidden unless --all
	}
	defer withFakes(t, mounts, stats)()

	// Without --all: /proc is hidden.
	out, _, err := run(t, "--output=source,target")
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if strings.Contains(out, "/proc") {
		t.Errorf("/proc should be hidden without --all:\n%s", out)
	}

	// With --all: /proc is shown.
	outAll, _, err := run(t, "--all", "--output=source,target")
	if err != nil {
		t.Fatalf("Run -a error: %v", err)
	}
	if !strings.Contains(outAll, "/proc") {
		t.Errorf("/proc should appear with --all:\n%s", outAll)
	}
}

func TestRunDefaultBehaviorUnchanged(t *testing.T) {
	// No GNU flags, no operands: must stay cwd-only classic layout.
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
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("default df should print header + 1 row, got %d lines:\n%s", len(lines), out)
	}
	if !strings.HasPrefix(lines[0], "Filesystem") || !strings.Contains(lines[0], "1K-blocks") {
		t.Errorf("default header changed: %q", lines[0])
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
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q:\n%s", want, out)
		}
	}
}
