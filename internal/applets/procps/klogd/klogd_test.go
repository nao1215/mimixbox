package klogd

import (
	"bytes"
	"context"
	"errors"
	"log/syslog"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type forwarded struct {
	prio syslog.Priority
	msg  string
}

func stub(t *testing.T, log string, readErr, sendErr error) *[]forwarded {
	t.Helper()
	var sent []forwarded
	or, ol := readKernelLog, logFunc
	readKernelLog = func() ([]byte, error) {
		if readErr != nil {
			return nil, readErr
		}
		return []byte(log), nil
	}
	logFunc = func(p syslog.Priority, msg string) error {
		if sendErr != nil {
			return sendErr
		}
		sent = append(sent, forwarded{p, msg})
		return nil
	}
	t.Cleanup(func() { readKernelLog, logFunc = or, ol })
	return &sent
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestForwards(t *testing.T) {
	sent := stub(t, "<6>kernel up\n<3>disk error\n\n", nil, nil)
	if err := run(t, "-o"); err != nil {
		t.Fatal(err)
	}
	if len(*sent) != 2 {
		t.Fatalf("forwarded %d, want 2: %+v", len(*sent), *sent)
	}
	if (*sent)[0].msg != "kernel up" || (*sent)[0].prio != syslog.LOG_KERN|syslog.LOG_INFO {
		t.Errorf("first = %+v", (*sent)[0])
	}
	if (*sent)[1].msg != "disk error" || (*sent)[1].prio != syslog.LOG_KERN|syslog.LOG_ERR {
		t.Errorf("second = %+v", (*sent)[1])
	}
}

func TestSplitPriority(t *testing.T) {
	t.Parallel()
	if p, m := splitPriority("<4>warn msg"); p != syslog.LOG_KERN|syslog.LOG_WARNING || m != "warn msg" {
		t.Errorf("split = %d, %q", p, m)
	}
	if p, m := splitPriority("no prefix"); p != syslog.LOG_KERN|syslog.LOG_INFO || m != "no prefix" {
		t.Errorf("split default = %d, %q", p, m)
	}
}

func TestReadFailure(t *testing.T) {
	stub(t, "", errors.New("permission denied"), nil)
	if err := run(t, "-o"); err == nil {
		t.Errorf("a read failure should fail")
	}
}

func TestForwardFailure(t *testing.T) {
	stub(t, "<6>line\n", nil, errors.New("no syslogd"))
	if err := run(t, "-o"); err == nil {
		t.Errorf("a forward failure should fail")
	}
}
