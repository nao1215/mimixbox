package nsenter

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withStub(t *testing.T, err error) *[]string {
	t.Helper()
	var entered []string
	orig := setnsFn
	setnsFn = func(path string) error {
		entered = append(entered, path)
		return err
	}
	t.Cleanup(func() { setnsFn = orig })
	return &entered
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestEntersNamespaceAndRuns(t *testing.T) {
	entered := withStub(t, nil)
	out, err := run(t, "-t", "1234", "-n", "echo", "hi")
	if err != nil {
		t.Fatal(err)
	}
	if len(*entered) != 1 || (*entered)[0] != "/proc/1234/ns/net" {
		t.Errorf("entered = %v", *entered)
	}
	if out != "hi" {
		t.Errorf("output = %q", out)
	}
}

func TestEntersMultiple(t *testing.T) {
	entered := withStub(t, nil)
	if _, err := run(t, "-t", "5", "-m", "-u", "true"); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"/proc/5/ns/mnt": true, "/proc/5/ns/uts": true}
	if len(*entered) != 2 || !want[(*entered)[0]] || !want[(*entered)[1]] {
		t.Errorf("entered = %v, want mnt and uts of PID 5", *entered)
	}
}

func TestRequiresTarget(t *testing.T) {
	withStub(t, nil)
	if _, err := run(t, "-n", "echo", "x"); err == nil {
		t.Errorf("missing -t should fail")
	}
}

func TestRequiresNamespace(t *testing.T) {
	withStub(t, nil)
	if _, err := run(t, "-t", "1", "echo", "x"); err == nil {
		t.Errorf("no namespace flag should fail")
	}
}

func TestSetnsFailure(t *testing.T) {
	withStub(t, errors.New("permission denied"))
	if _, err := run(t, "-t", "1", "-n", "echo", "x"); err == nil {
		t.Errorf("a setns failure should fail")
	}
}

func TestExitCodePropagates(t *testing.T) {
	withStub(t, nil)
	_, err := run(t, "-t", "1", "-u", "sh", "-c", "exit 7")
	ee, ok := err.(*command.ExitError)
	if !ok || ee.Code != 7 {
		t.Errorf("err = %v, want exit 7", err)
	}
}
