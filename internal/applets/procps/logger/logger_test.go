package logger

import (
	"bytes"
	"context"
	"errors"
	"log/syslog"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type logged struct {
	prio syslog.Priority
	tag  string
	msg  string
}

func withStub(t *testing.T) *logged {
	t.Helper()
	captured := &logged{}
	orig := logFunc
	logFunc = func(p syslog.Priority, tag, msg string) error {
		*captured = logged{p, tag, msg}
		return nil
	}
	t.Cleanup(func() { logFunc = orig })
	return captured
}

func run(t *testing.T, in string, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(in), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestLogsMessage(t *testing.T) {
	got := withStub(t)
	if err := run(t, "", "-t", "myapp", "hello", "world"); err != nil {
		t.Fatal(err)
	}
	if got.tag != "myapp" || got.msg != "hello world" {
		t.Errorf("logged = %+v", *got)
	}
	if got.prio != syslog.LOG_USER|syslog.LOG_NOTICE {
		t.Errorf("default priority = %d, want user.notice", got.prio)
	}
}

func TestPriority(t *testing.T) {
	got := withStub(t)
	if err := run(t, "", "-p", "auth.warning", "x"); err != nil {
		t.Fatal(err)
	}
	if got.prio != syslog.LOG_AUTH|syslog.LOG_WARNING {
		t.Errorf("priority = %d, want auth.warning", got.prio)
	}
}

func TestMessageFromStdin(t *testing.T) {
	got := withStub(t)
	if err := run(t, "from stdin\n", "-t", "x"); err != nil {
		t.Fatal(err)
	}
	if got.msg != "from stdin" {
		t.Errorf("stdin msg = %q", got.msg)
	}
}

// errReader fails on the first read, modeling an unreadable stdin.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("is a directory") }

func TestStdinReadErrorIsReported(t *testing.T) {
	withStub(t)
	io := command.IO{In: errReader{}, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	// No message operand: logger reads the message from stdin. An unreadable
	// stdin must fail instead of logging an empty message.
	if err := New().Run(context.Background(), io, []string{"-t", "x"}); err == nil {
		t.Fatal("logger must report an unreadable stdin instead of treating it as empty")
	}
}

func TestInvalidPriority(t *testing.T) {
	withStub(t)
	if err := run(t, "", "-p", "bogus.level", "x"); err == nil {
		t.Errorf("invalid facility should fail")
	}
	if err := run(t, "", "-p", "user.bogus", "x"); err == nil {
		t.Errorf("invalid level should fail")
	}
}

func TestParsePriority(t *testing.T) {
	t.Parallel()
	if p, err := parsePriority("daemon.err"); err != nil || p != syslog.LOG_DAEMON|syslog.LOG_ERR {
		t.Errorf("daemon.err = %d, %v", p, err)
	}
	// A bare level defaults to the user facility.
	if p, err := parsePriority("info"); err != nil || p != syslog.LOG_USER|syslog.LOG_INFO {
		t.Errorf("info = %d, %v", p, err)
	}
}
