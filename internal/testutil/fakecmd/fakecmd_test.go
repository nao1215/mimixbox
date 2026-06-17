package fakecmd

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runFake resolves name on the (already prepared) PATH and runs it, returning
// combined stdout and the exit code.
func runFake(t *testing.T, dir, name string, args ...string) (string, int) {
	t.Helper()
	bin := filepath.Join(dir, name)
	cmd := exec.Command(bin, args...)
	out, err := cmd.Output()
	code := 0
	if err != nil {
		var ee *exec.ExitError
		if as, ok := err.(*exec.ExitError); ok {
			ee = as
		}
		if ee != nil {
			code = ee.ExitCode()
		} else {
			t.Fatalf("running %s: %v", name, err)
		}
	}
	return string(out), code
}

func TestEcho(t *testing.T) {
	dir := Dir(t, "echo")
	if out, code := runFake(t, dir, "echo", "hello", "world"); out != "hello world\n" || code != 0 {
		t.Errorf("echo = %q (%d)", out, code)
	}
	if out, _ := runFake(t, dir, "echo", "-n", "x"); out != "x" {
		t.Errorf("echo -n = %q", out)
	}
}

func TestTrueFalse(t *testing.T) {
	dir := Dir(t, "true", "false")
	if _, code := runFake(t, dir, "true"); code != 0 {
		t.Errorf("true exit = %d", code)
	}
	if _, code := runFake(t, dir, "false"); code != 1 {
		t.Errorf("false exit = %d", code)
	}
}

func TestPrintf(t *testing.T) {
	dir := Dir(t, "printf")
	if out, _ := runFake(t, dir, "printf", "%s-%s", "a", "b"); out != "a-b" {
		t.Errorf("printf = %q", out)
	}
	if out, _ := runFake(t, dir, "printf", "hello"); out != "hello" {
		t.Errorf("printf literal = %q", out)
	}
	if out, _ := runFake(t, dir, "printf", `one\ntwo\n`); out != "one\ntwo\n" {
		t.Errorf("printf escape = %q", out)
	}
}

func TestWc(t *testing.T) {
	dir := Dir(t, "wc")
	cmd := exec.Command(filepath.Join(dir, "wc"), "-c")
	cmd.Stdin = strings.NewReader("foo")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != "3" {
		t.Errorf("wc -c = %q", out)
	}
}

func TestShExit(t *testing.T) {
	dir := Dir(t, "sh")
	if _, code := runFake(t, dir, "sh", "-c", "exit 3"); code != 3 {
		t.Errorf("sh exit code = %d", code)
	}
}

func TestUseSetsPath(t *testing.T) {
	Use(t, "echo")
	if _, err := exec.LookPath("echo"); err != nil {
		t.Errorf("echo should resolve after Use: %v", err)
	}
}
