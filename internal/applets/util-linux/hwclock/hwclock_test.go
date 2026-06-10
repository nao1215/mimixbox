package hwclock

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func withRTC(t *testing.T, when time.Time, err error) {
	t.Helper()
	orig := readRTC
	readRTC = func() (time.Time, error) { return when, err }
	t.Cleanup(func() { readRTC = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestShowLocal(t *testing.T) {
	rtc := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	withRTC(t, rtc, nil)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	want := rtc.Local().Format("2006-01-02 15:04:05")
	if out != want {
		t.Errorf("hwclock = %q, want %q", out, want)
	}
}

func TestShowUTC(t *testing.T) {
	rtc := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	withRTC(t, rtc, nil)
	out, err := run(t, "-u")
	if err != nil {
		t.Fatal(err)
	}
	if out != "2026-06-10 12:00:00" {
		t.Errorf("hwclock -u = %q, want 2026-06-10 12:00:00", out)
	}
}

func TestReadFailure(t *testing.T) {
	withRTC(t, time.Time{}, errors.New("permission denied"))
	if _, err := run(t); err == nil {
		t.Errorf("an RTC read failure should fail")
	}
}
