package switchroot

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withStub(t *testing.T, switchErr error) (*string, *[]string) {
	t.Helper()
	switched := new(string)
	*switched = "<unset>"
	var execArgv []string
	osw, oex := switchFn, execFn
	switchFn = func(newRoot string) error {
		*switched = newRoot
		return switchErr
	}
	execFn = func(path string, argv []string) error {
		execArgv = append([]string{path}, argv...)
		return nil // pretend the exec succeeded without replacing the process
	}
	t.Cleanup(func() { switchFn, execFn = osw, oex })
	return switched, &execArgv
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestSwitchesAndExecs(t *testing.T) {
	switched, execArgv := withStub(t, nil)
	dir := t.TempDir()
	if err := run(t, dir, "/sbin/init", "single"); err != nil {
		t.Fatal(err)
	}
	if *switched != dir {
		t.Errorf("switched to %q, want %q", *switched, dir)
	}
	// execFn receives (path, argv) -> our stub records path + argv.
	if len(*execArgv) != 3 || (*execArgv)[1] != "/sbin/init" || (*execArgv)[2] != "single" {
		t.Errorf("exec argv = %v", *execArgv)
	}
}

func TestRequiresNewRootAndInit(t *testing.T) {
	withStub(t, nil)
	dir := t.TempDir()
	if err := run(t, dir); err == nil {
		t.Errorf("missing init should fail")
	}
	if err := run(t); err == nil {
		t.Errorf("no arguments should fail")
	}
}

func TestNewRootMustBeDirectory(t *testing.T) {
	withStub(t, nil)
	if err := run(t, "/no/such/dir", "/init"); err == nil {
		t.Errorf("a non-directory NEW_ROOT should fail")
	}
}

func TestSwitchFailure(t *testing.T) {
	withStub(t, errors.New("operation not permitted"))
	dir := t.TempDir()
	if err := run(t, dir, "/sbin/init"); err == nil {
		t.Errorf("a switch failure should fail")
	}
}
