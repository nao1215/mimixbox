package showkey

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args []string) (string, string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestSelectMode(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name                      string
		ascii, keycodes, scancode bool
		want                      mode
		wantErr                   bool
	}{
		{"default", false, false, false, modeKeycode, false},
		{"keycode", false, true, false, modeKeycode, false},
		{"ascii", true, false, false, modeASCII, false},
		{"scancode", false, false, true, modeScancode, false},
		{"conflict-a-s", true, false, true, modeKeycode, true},
		{"conflict-all", true, true, true, modeKeycode, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectMode(tc.ascii, tc.keycodes, tc.scancode)
			if tc.wantErr != (err != nil) {
				t.Fatalf("selectMode err = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Errorf("selectMode = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestModeString(t *testing.T) {
	t.Parallel()
	if modeASCII.String() != "ASCII codes" || modeScancode.String() != "scancodes" || modeKeycode.String() != "keycodes" {
		t.Error("unexpected mode strings")
	}
}

// TestRunCapabilityError asserts that without a real console showkey fails
// deterministically rather than silently succeeding.
func TestRunCapabilityError(t *testing.T) {
	t.Parallel()
	_, errBuf, err := run(t, nil)
	if err == nil {
		t.Fatal("expected a capability error without a console")
	}
	_ = errBuf
}

func TestRunConflict(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, []string{"-a", "-s"}); err == nil {
		t.Error("expected error for conflicting modes")
	}
}

func TestRunUnexpectedArg(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, []string{"foo"}); err == nil {
		t.Error("expected error for unexpected argument")
	}
}

// TestRunInjectedSuccess exercises the narrow success path through the injected
// interactive function.
func TestRunInjectedSuccess(t *testing.T) {
	orig := interactiveFn
	interactiveFn = func(_ context.Context, _ command.IO, _ mode) error { return nil }
	defer func() { interactiveFn = orig }()
	if _, _, err := run(t, []string{"-k"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, []string{"--help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Usage: showkey") {
		t.Errorf("help missing usage:\n%s", out)
	}
}
