package stty

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// withFakeTerminal points stdin at a real *os.File (so fdOf succeeds) and stubs
// the termios get/set so no actual terminal is touched. It returns the captured
// termios that a set would write.
func withFakeTerminal(t *testing.T, initial *unix.Termios) (*os.File, *unix.Termios) {
	t.Helper()
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })

	captured := &unix.Termios{}
	origGet, origSet := getTermios, setTermios
	getTermios = func(int) (*unix.Termios, error) { cp := *initial; return &cp, nil }
	setTermios = func(_ int, tm *unix.Termios) error { *captured = *tm; return nil }
	t.Cleanup(func() { getTermios, setTermios = origGet, origSet })
	return f, captured
}

func TestNotATty(t *testing.T) {
	t.Parallel()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Errorf("a non-terminal stdin should fail")
	}
}

func TestPrintAll(t *testing.T) {
	in, _ := withFakeTerminal(t, &unix.Termios{Lflag: unix.ECHO | unix.ICANON | unix.ISIG, Ospeed: unix.B38400})
	out := &bytes.Buffer{}
	io := command.IO{In: in, Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"-a"}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "speed 38400 baud") || !strings.Contains(s, "echo") || !strings.Contains(s, "-echoe") {
		t.Errorf("stty -a = %q", s)
	}
}

func TestSetDisableEcho(t *testing.T) {
	in, captured := withFakeTerminal(t, &unix.Termios{Lflag: unix.ECHO | unix.ICANON, Ospeed: unix.B38400})
	io := command.IO{In: in, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"-echo"}); err != nil {
		t.Fatal(err)
	}
	if captured.Lflag&unix.ECHO != 0 {
		t.Errorf("ECHO should be cleared, lflag = %x", captured.Lflag)
	}
	if captured.Lflag&unix.ICANON == 0 {
		t.Errorf("ICANON should remain set")
	}
}

func TestApplySettings(t *testing.T) {
	t.Parallel()
	tm := &unix.Termios{}
	if err := applySettings(tm, []string{"echo", "icanon"}); err != nil {
		t.Fatal(err)
	}
	if tm.Lflag&unix.ECHO == 0 || tm.Lflag&unix.ICANON == 0 {
		t.Errorf("echo/icanon not set: %x", tm.Lflag)
	}
	if err := applySettings(tm, []string{"raw"}); err != nil {
		t.Fatal(err)
	}
	if tm.Lflag&unix.ICANON != 0 || tm.Lflag&unix.ECHO != 0 {
		t.Errorf("raw should clear icanon/echo: %x", tm.Lflag)
	}
	if err := applySettings(tm, []string{"bogus"}); err == nil {
		t.Errorf("an invalid setting should fail")
	}
}
