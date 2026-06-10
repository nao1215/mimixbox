package dmesg

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withLog(t *testing.T, data string, err error) {
	t.Helper()
	orig := readKernelLog
	readKernelLog = func() ([]byte, error) {
		if err != nil {
			return nil, err
		}
		return []byte(data), nil
	}
	t.Cleanup(func() { readKernelLog = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestStripsPriority(t *testing.T) {
	withLog(t, "<6>[    0.000000] boot\n<4>[    1.234567] warning\n", nil)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	want := "[    0.000000] boot\n[    1.234567] warning\n"
	if out != want {
		t.Errorf("dmesg = %q, want %q", out, want)
	}
}

func TestRawKeepsPrefix(t *testing.T) {
	raw := "<6>[    0.000000] boot\n"
	withLog(t, raw, nil)
	out, err := run(t, "-r")
	if err != nil {
		t.Fatal(err)
	}
	if out != raw {
		t.Errorf("dmesg -r = %q, want %q", out, raw)
	}
}

func TestReadFailure(t *testing.T) {
	withLog(t, "", errors.New("operation not permitted"))
	if _, err := run(t); err == nil {
		t.Errorf("a read failure should fail")
	}
}

func TestStripPriority(t *testing.T) {
	t.Parallel()
	if got := stripPriority("<6>[ 0.0] hi"); got != "[ 0.0] hi" {
		t.Errorf("stripPriority = %q", got)
	}
	if got := stripPriority("no prefix"); got != "no prefix" {
		t.Errorf("stripPriority(no prefix) = %q", got)
	}
}
