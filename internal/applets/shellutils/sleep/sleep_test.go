package sleep

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestParseDuration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		want    time.Duration
		wantErr bool
	}{
		{"bare number is seconds", []string{"1"}, time.Second, false},
		{"seconds suffix", []string{"2s"}, 2 * time.Second, false},
		{"minutes suffix", []string{"3m"}, 3 * time.Minute, false},
		{"hours suffix", []string{"1h"}, time.Hour, false},
		{"days suffix", []string{"1d"}, 24 * time.Hour, false},
		{"multiple operands summed", []string{"1", "2s", "1m"}, time.Minute + 3*time.Second, false},
		{"zero", []string{"0"}, 0, false},
		{"invalid", []string{"abc"}, 0, true},
		{"invalid suffix value", []string{"xs"}, 0, true},
		{"empty", []string{""}, 0, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseDuration(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseDuration(%v) expected error", tt.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseDuration(%v) unexpected error = %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("parseDuration(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunZeroReturnsQuickly(t *testing.T) {
	t.Parallel()
	done := make(chan error, 1)
	go func() {
		_, _, err := run(t, "0")
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run(\"0\") error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run(\"0\") did not return quickly")
	}
}

func TestRunInvalidOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "abc")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "invalid time interval 'abc'") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "missing operand") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunContextCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(ctx, io, []string{"100"})
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
