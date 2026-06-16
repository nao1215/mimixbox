package date_test

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/date"
	"github.com/nao1215/mimixbox/internal/command"
)

// fixed is the deterministic clock used across the tests: 2023-11-14T22:13:20Z,
// whose Unix time is the round number 1700000000.
var fixed = time.Date(2023, time.November, 14, 22, 13, 20, 0, time.UTC)

// nowMu serializes access to the package-level clock so the tests that swap it
// (via date.SetNow) do not race when the Go test runner schedules them
// concurrently with -parallel.
var nowMu sync.Mutex

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	nowMu.Lock()
	defer nowMu.Unlock()
	restore := date.SetNow(fixed)
	defer restore()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := date.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNewAndMetadata(t *testing.T) {
	t.Parallel()
	c := date.New()
	if c == nil {
		t.Fatal("New returned nil")
	}
	if c.Name() != "date" {
		t.Errorf("Name = %q, want date", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis is empty")
	}
}

func TestRunDefaultNonEmpty(t *testing.T) {
	t.Parallel()
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("default output is empty")
	}
	if !strings.Contains(out, "2023") {
		t.Errorf("default output = %q, want it to mention the year", out)
	}
}

func TestRunFormats(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"iso date", []string{"-u", "+%Y-%m-%d"}, "2023-11-14"},
		{"time utc", []string{"-u", "+%H:%M:%S"}, "22:13:20"},
		{"epoch", []string{"+%s"}, "1700000000"},
		{"percent", []string{"+%%"}, "%"},
		{"iso -u -F", []string{"-u", "+%F"}, "2023-11-14"},
		{"iso -I flag", []string{"-u", "-I"}, "2023-11-14"},
		{"iso seconds", []string{"-u", "--iso-8601=seconds"}, "2023-11-14T22:13:20+00:00"},
		{"rfc email", []string{"-u", "-R"}, "Tue, 14 Nov 2023 22:13:20 +0000"},
		{"weekday abbr", []string{"-u", "+%a %A"}, "Tue Tuesday"},
		{"month names", []string{"-u", "+%b %B"}, "Nov November"},
		{"combined T", []string{"-u", "+%T"}, "22:13:20"},
		{"12h pm", []string{"-u", "+%I %p"}, "10 PM"},
		{"yearday", []string{"-u", "+%j"}, "318"},
		{"two digit year", []string{"-u", "+%y"}, "23"},
		{"literal text", []string{"-u", "+year=%Y"}, "year=2023"},
		{"date string epoch", []string{"-u", "-d", "@0", "+%Y-%m-%d"}, "1970-01-01"},
		{"date string rfc3339", []string{"-u", "-d", "2000-01-02T03:04:05Z", "+%T"}, "03:04:05"},
		{"date string ymd", []string{"-u", "-d", "1999-12-31", "+%F"}, "1999-12-31"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, errOut, err := run(t, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
			}
			if got := strings.TrimRight(out, "\n"); got != tt.want {
				t.Errorf("out = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestHour12Midnight covers hour12's h==0 -> 12 branch (midnight in 12-hour
// clock is 12 AM) and the %P lowercase am rendering.
func TestHour12Midnight(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "-u", "-d", "2023-11-14T00:00:00Z", "+%I %p %P")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if got := strings.TrimRight(out, "\n"); got != "12 AM am" {
		t.Errorf("out = %q, want %q", got, "12 AM am")
	}
}

// TestSundayISOWeekday covers specifier's %u Sunday=7 special case.
func TestSundayISOWeekday(t *testing.T) {
	t.Parallel()
	// 2023-11-12 is a Sunday.
	out, errOut, err := run(t, "-u", "-d", "2023-11-12", "+%u %w")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if got := strings.TrimRight(out, "\n"); got != "7 0" {
		t.Errorf("out = %q, want %q (ISO Sunday=7, w Sunday=0)", got, "7 0")
	}
}

// TestISOFormats covers the hours/minutes/ns granularities of --iso-8601.
func TestISOFormats(t *testing.T) {
	t.Parallel()
	tests := []struct {
		arg  string
		want string
	}{
		{"hours", "2023-11-14T22+00:00"},
		{"minutes", "2023-11-14T22:13+00:00"},
		{"ns", "2023-11-14T22:13:20+00:00"}, // zero nanoseconds are trimmed by the ,999... layout
		{"date", "2023-11-14"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.arg, func(t *testing.T) {
			out, errOut, err := run(t, "-u", "--iso-8601="+tt.arg)
			if err != nil {
				t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
			}
			if got := strings.TrimRight(out, "\n"); got != tt.want {
				t.Errorf("--iso-8601=%s out = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}

// TestInvalidISOArg covers isoFormat's default (error) branch.
func TestInvalidISOArg(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "--iso-8601=bogus")
	if err == nil {
		t.Fatal("expected error for invalid --iso-8601 argument")
	}
	if !strings.Contains(errOut, "invalid argument") {
		t.Errorf("stderr = %q, want invalid argument message", errOut)
	}
}

// TestExtraOperand covers operandFormat's "extra operand" branch.
func TestExtraOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "+%Y", "+%m")
	if err == nil {
		t.Fatal("expected error for a second operand")
	}
	if !strings.Contains(errOut, "extra operand") {
		t.Errorf("stderr = %q, want extra operand message", errOut)
	}
}

func TestRunInvalidDate(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "-d", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
	if !strings.Contains(errOut, "date:") {
		t.Errorf("stderr = %q, want date error prefix", errOut)
	}
}

func TestRunInvalidOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "%Y")
	if err == nil {
		t.Fatal("expected error for operand without leading +")
	}
	if !strings.Contains(errOut, "date:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunSetUnsupported(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "-s", "2020-01-01")
	if err == nil {
		t.Fatal("expected not-supported error for -s")
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: date") {
		t.Errorf("--help out = %q", out)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q:\n%s", want, out)
		}
	}

	out, _, err = run(t, "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "date (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}

func TestFormatTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		format string
		want   string
	}{
		{"%Y-%m-%d", "2023-11-14"},
		{"%H:%M:%S", "22:13:20"},
		{"%s", "1700000000"},
		{"%%", "%"},
		{"%F", "2023-11-14"},
		{"%T", "22:13:20"},
		{"%A", "Tuesday"},
		{"%a", "Tue"},
		{"%B", "November"},
		{"%b", "Nov"},
		{"%j", "318"},
		{"%y", "23"},
		{"%C", "20"},
		{"%p", "PM"},
		{"%P", "pm"},
		{"%I", "10"},
		{"%e", "14"},
		{"%m/%d/%y", "11/14/23"},
		{"%R", "22:13"},
		{"%u", "2"},
		{"%w", "2"},
		{"%z", "+0000"},
		{"%Z", "UTC"},
		{"literal", "literal"},
		{"a%Yb", "a2023b"},
		{"%n", "\n"},
		{"%t", "\t"},
		{"trailing %", "trailing %"},
		{"%Q", "%Q"},                        // unknown specifier passes through
		{"%c", "Tue Nov  14 22:13:20 2023"}, // the layout pads the day field with a leading space
		{"%D", "11/14/23"},
		{"%r", "10:13:20 PM"},
		{"%x", "11/14/23"},
		{"%X", "22:13:20"},
		{"%h", "Nov"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.format, func(t *testing.T) {
			t.Parallel()
			if got := date.FormatTime(fixed, tt.format); got != tt.want {
				t.Errorf("FormatTime(%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}
