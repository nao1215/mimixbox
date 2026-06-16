package uname

import (
	"errors"
	"strings"
	"testing"
)

// TestEachIndividualFlag exercises every per-field flag, covering the -n, -r
// and -v branches that the existing tests do not select individually.
func TestEachIndividualFlag(t *testing.T) {
	tests := []struct {
		flag string
		want string
	}{
		{"-n", "host\n"},
		{"-r", "6.6.0\n"},
		{"-v", "#1 SMP\n"},
		{"-s", "Linux\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.flag, func(t *testing.T) {
			withStub(t)
			out, _, err := run(t, tt.flag)
			if err != nil {
				t.Fatalf("Run %s error = %v", tt.flag, err)
			}
			if out != tt.want {
				t.Errorf("%s -> out = %q, want %q", tt.flag, out, tt.want)
			}
		})
	}
}

// TestAllFieldsCombineInOrder verifies a multi-flag invocation prints the
// fields in their fixed order, regardless of flag order on the command line.
func TestAllFieldsCombineInOrder(t *testing.T) {
	withStub(t)
	out, _, err := run(t, "-v", "-n", "-r")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// Output order follows the field order in Run, not the argument order.
	if out != "host 6.6.0 #1 SMP\n" {
		t.Errorf("out = %q, want %q", out, "host 6.6.0 #1 SMP\n")
	}
}

// TestSysInfoErrorIsReported covers the error branch of Run when the underlying
// system call fails.
func TestSysInfoErrorIsReported(t *testing.T) {
	orig := sysInfo
	sysInfo = func() (info, error) { return info{}, errors.New("boom") }
	t.Cleanup(func() { sysInfo = orig })

	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected an error when sysInfo fails")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("err = %v, want it to mention the underlying failure", err)
	}
}

// TestCharsToString covers the NUL-termination logic, including a string that
// fills the whole buffer (no NUL) and one terminated early.
func TestCharsToString(t *testing.T) {
	t.Parallel()
	if got := charsToString([]byte{'a', 'b', 'c', 0, 'd'}); got != "abc" {
		t.Errorf("charsToString stopped at NUL = %q, want abc", got)
	}
	if got := charsToString([]byte{'x', 'y', 'z'}); got != "xyz" {
		t.Errorf("charsToString without NUL = %q, want xyz", got)
	}
	if got := charsToString([]byte{}); got != "" {
		t.Errorf("charsToString empty = %q, want empty", got)
	}
}
