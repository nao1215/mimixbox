package signal_test

import (
	"errors"
	"os"
	osignal "os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/signal"
)

// sendSelf sends sig to the current process.
func sendSelf(t *testing.T, sig syscall.Signal) {
	t.Helper()
	if err := syscall.Kill(syscall.Getpid(), sig); err != nil {
		t.Fatalf("kill self with %v: %v", sig, err)
	}
}

// deliversAfter installs a Notify handler for sig, sends sig to the current
// process, and reports whether it is delivered within a short window. A
// delivered signal means the prior disposition (not SIG_IGN) is in effect;
// SIG_IGN would suppress delivery to the Notify channel.
func deliversAfter(t *testing.T, sig syscall.Signal) bool {
	t.Helper()
	ch := make(chan os.Signal, 1)
	osignal.Notify(ch, sig)
	defer osignal.Stop(ch)

	sendSelf(t, sig)
	select {
	case <-ch:
		return true
	case <-time.After(300 * time.Millisecond):
		return false
	}
}

func TestIgnoreDuringRunsFunc(t *testing.T) {
	called := false
	err := signal.IgnoreDuring([]os.Signal{syscall.SIGUSR1}, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("IgnoreDuring returned error: %v", err)
	}
	if !called {
		t.Fatal("IgnoreDuring did not invoke fn")
	}
}

func TestIgnoreDuringIgnoresSignalWhileRunning(t *testing.T) {
	// A Notify handler installed before IgnoreDuring must NOT receive the
	// signal while IgnoreDuring holds SIG_IGN: signal.Ignore overrides the
	// Notify disposition for the duration.
	ch := make(chan os.Signal, 1)
	osignal.Notify(ch, syscall.SIGUSR1)
	defer osignal.Stop(ch)

	err := signal.IgnoreDuring([]os.Signal{syscall.SIGUSR1}, func() error {
		sendSelf(t, syscall.SIGUSR1)
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("IgnoreDuring returned error: %v", err)
	}
	select {
	case <-ch:
		t.Fatal("signal was delivered while IgnoreDuring held SIG_IGN")
	default:
	}
}

func TestIgnoreDuringPropagatesError(t *testing.T) {
	sentinel := errors.New("boom")
	err := signal.IgnoreDuring([]os.Signal{syscall.SIGUSR1}, func() error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("IgnoreDuring error = %v, want %v", err, sentinel)
	}
}

func TestIgnoreDuringRestoresOnSuccess(t *testing.T) {
	if err := signal.IgnoreDuring([]os.Signal{syscall.SIGUSR1}, func() error {
		return nil
	}); err != nil {
		t.Fatalf("IgnoreDuring returned error: %v", err)
	}
	if !deliversAfter(t, syscall.SIGUSR1) {
		t.Fatal("signal was still ignored after successful IgnoreDuring")
	}
}

func TestIgnoreDuringRestoresOnError(t *testing.T) {
	_ = signal.IgnoreDuring([]os.Signal{syscall.SIGUSR1}, func() error {
		return errors.New("boom")
	})
	if !deliversAfter(t, syscall.SIGUSR1) {
		t.Fatal("signal was still ignored after IgnoreDuring returned an error")
	}
}

func TestIgnoreDuringRestoresOnPanic(t *testing.T) {
	func() {
		defer func() { _ = recover() }()
		_ = signal.IgnoreDuring([]os.Signal{syscall.SIGUSR1}, func() error {
			panic("boom")
		})
	}()
	if !deliversAfter(t, syscall.SIGUSR1) {
		t.Fatal("signal was still ignored after IgnoreDuring panicked")
	}
}

func TestIgnoreDuringNoSignals(t *testing.T) {
	called := false
	err := signal.IgnoreDuring(nil, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("IgnoreDuring(nil) returned error: %v", err)
	}
	if !called {
		t.Fatal("IgnoreDuring(nil) did not invoke fn")
	}
}
