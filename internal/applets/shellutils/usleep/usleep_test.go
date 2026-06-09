package usleep

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestUsleepDuration(t *testing.T) {
	var got time.Duration
	orig := sleep
	sleep = func(d time.Duration) { got = d }
	defer func() { sleep = orig }()

	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"1500"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if got != 1500*time.Microsecond {
		t.Errorf("slept %v, want 1500µs", got)
	}
}

func TestUsleepNoArg(t *testing.T) {
	var got time.Duration
	orig := sleep
	sleep = func(d time.Duration) { got = d }
	defer func() { sleep = orig }()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if got != 0 {
		t.Errorf("no arg should sleep 0, got %v", got)
	}
}

func TestUsleepInvalid(t *testing.T) {
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"-5"}); err == nil {
		t.Errorf("negative value should fail")
	}
}
