package setsid

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestSetsidRunsProgram(t *testing.T) {
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skipf("echo not on PATH: %v", err)
	}
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"echo", "hello"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Errorf("output = %q", out.String())
	}
}

func TestSetsidPropagatesExit(t *testing.T) {
	if _, err := exec.LookPath("false"); err != nil {
		t.Skipf("false not on PATH: %v", err)
	}
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, []string{"false"})
	var ee *command.ExitError
	if err == nil {
		t.Fatal("false should produce a non-zero exit")
	}
	if e, ok := err.(*command.ExitError); ok {
		ee = e
	}
	if ee == nil || ee.Code != 1 {
		t.Errorf("err = %v, want exit 1", err)
	}
}

func TestSetsidMissingProgram(t *testing.T) {
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Errorf("missing program should fail")
	}
}
