package cal_test

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/cal"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	stdio := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := cal.New().Run(context.Background(), stdio, args)
	return out.String(), errBuf.String(), err
}

func TestNewAndMetadata(t *testing.T) {
	t.Parallel()
	c := cal.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Name() != "cal" {
		t.Errorf("Name() = %q, want cal", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestMonthNovember2023Sunday(t *testing.T) {
	t.Parallel()
	want := "   November 2023\n" +
		"Su Mo Tu We Th Fr Sa\n" +
		"          1  2  3  4\n" +
		" 5  6  7  8  9 10 11\n" +
		"12 13 14 15 16 17 18\n" +
		"19 20 21 22 23 24 25\n" +
		"26 27 28 29 30\n"
	if got := cal.Month(2023, time.November, false); got != want {
		t.Errorf("Month() mismatch.\n got:\n%q\nwant:\n%q", got, want)
	}
}

func TestMonthNovember2023Monday(t *testing.T) {
	t.Parallel()
	want := "   November 2023\n" +
		"Mo Tu We Th Fr Sa Su\n" +
		"       1  2  3  4  5\n" +
		" 6  7  8  9 10 11 12\n" +
		"13 14 15 16 17 18 19\n" +
		"20 21 22 23 24 25 26\n" +
		"27 28 29 30\n"
	if got := cal.Month(2023, time.November, true); got != want {
		t.Errorf("Month() mismatch.\n got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRunMonthYearOperands(t *testing.T) {
	t.Parallel()
	want := "   November 2023\n" +
		"Su Mo Tu We Th Fr Sa\n" +
		"          1  2  3  4\n" +
		" 5  6  7  8  9 10 11\n" +
		"12 13 14 15 16 17 18\n" +
		"19 20 21 22 23 24 25\n" +
		"26 27 28 29 30\n"
	out, errOut, err := run(t, "11", "2023")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if out != want {
		t.Errorf("Run out = %q, want %q", out, want)
	}
}

func TestRunMondayOption(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-m", "11", "2023")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.HasPrefix(out, "   November 2023\nMo Tu We Th Fr Sa Su\n") {
		t.Errorf("Run -m out = %q", out)
	}
}

func TestRunCurrentMonth(t *testing.T) {
	t.Parallel()
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	now := time.Now()
	header := now.Month().String()
	if !strings.Contains(out, header) {
		t.Errorf("current month out = %q, want to contain %q", out, header)
	}
	if !strings.Contains(out, "Su Mo Tu We Th Fr Sa") {
		t.Errorf("current month out missing weekday header: %q", out)
	}
}

func TestRunYearOperand(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "2023")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	for _, m := range []string{"January 2023", "December 2023"} {
		if !strings.Contains(out, m) {
			t.Errorf("year out missing %q:\n%s", m, out)
		}
	}
}

func TestRunYearOption(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-y")
	if err != nil {
		t.Fatalf("Run -y error = %v", err)
	}
	now := time.Now()
	if !strings.Contains(out, "January "+strconv.Itoa(now.Year())) {
		t.Errorf("year -y out missing January of current year:\n%s", out)
	}
}

func TestRunInvalidOperands(t *testing.T) {
	t.Parallel()
	tests := [][]string{
		{"abc"},
		{"13", "2023"},
		{"11", "notyear"},
		{"1", "2", "3"},
	}
	for _, args := range tests {
		args := args
		_, errOut, err := run(t, args...)
		if err == nil {
			t.Errorf("args %v: expected error", args)
		}
		if !strings.Contains(errOut, "cal:") {
			t.Errorf("args %v: stderr = %q, want cal: prefix", args, errOut)
		}
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: cal") {
		t.Errorf("--help out = %q", out)
	}

	out, _, err = run(t, "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "cal (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}
