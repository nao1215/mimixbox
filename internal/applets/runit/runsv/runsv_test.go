package runsv

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func stub(t *testing.T, counter *int32) {
	t.Helper()
	od, of := restartDelay, runOnceFn
	restartDelay = time.Millisecond
	runOnceFn = func(_ context.Context, _ string, _ command.IO) error {
		atomic.AddInt32(counter, 1)
		return nil
	}
	t.Cleanup(func() { restartDelay, runOnceFn = od, of })
}

func runAsync(ctx context.Context, dir string) <-chan error {
	done := make(chan error, 1)
	go func() {
		io := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		done <- New().Run(ctx, io, []string{dir})
	}()
	return done
}

func TestRestartsUntilCancelled(t *testing.T) {
	var count int32
	stub(t, &count)
	dir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	done := runAsync(ctx, dir)

	// Wait until ./run has been started several times (restart loop).
	deadline := time.Now().Add(2 * time.Second)
	for atomic.LoadInt32(&count) < 3 {
		if time.Now().After(deadline) {
			cancel()
			t.Fatalf("run started only %d times", atomic.LoadInt32(&count))
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runsv did not stop after cancellation")
	}

	// The supervise/ok file must have been created during supervision.
	if _, err := os.Stat(filepath.Join(dir, "supervise", "control")); err != nil {
		t.Errorf("control file not created: %v", err)
	}
}

func TestDownFileLeavesServiceStopped(t *testing.T) {
	var count int32
	stub(t, &count)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "down"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := runAsync(ctx, dir)
	time.Sleep(20 * time.Millisecond)
	if c := atomic.LoadInt32(&count); c != 0 {
		t.Errorf("a down service must not be started, ran %d times", c)
	}
	// It is still supervised (ok file exists).
	if _, err := os.Stat(filepath.Join(dir, "supervise", "ok")); err != nil {
		t.Errorf("a down service should still be supervised: %v", err)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runsv did not stop after cancellation")
	}
}

func TestNoDir(t *testing.T) {
	io := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Errorf("a missing directory should fail")
	}
}

func TestRejectsMissingDir(t *testing.T) {
	var count int32
	stub(t, &count)
	io := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"/no/such/service/dir"}); err == nil {
		t.Errorf("a nonexistent service directory should fail")
	}
}

func TestChildPidRecorded(t *testing.T) {
	// Use the real run step: a ./run that sleeps so we can read its pid.
	od := restartDelay
	restartDelay = time.Hour // the run blocks, so we only need one iteration
	t.Cleanup(func() { restartDelay = od })

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "run"), []byte("#!/bin/sh\nsleep 5\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := runAsync(ctx, dir)

	pidPath := filepath.Join(dir, "supervise", "pid")
	var pid string
	deadline := time.Now().Add(2 * time.Second)
	for {
		if data, err := os.ReadFile(pidPath); err == nil {
			if pid = strings.TrimSpace(string(data)); pid != "" {
				break
			}
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatal("supervise/pid was never written")
		}
		time.Sleep(2 * time.Millisecond)
	}
	cancel()
	<-done

	n, err := strconv.Atoi(pid)
	if err != nil || n == os.Getpid() {
		t.Errorf("pid = %q, want the child's pid (not the supervisor %d)", pid, os.Getpid())
	}
}
