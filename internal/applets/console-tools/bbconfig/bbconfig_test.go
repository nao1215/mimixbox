package bbconfig

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withApplets(t *testing.T, names []string, err error) {
	t.Helper()
	orig := appletNamesFn
	appletNamesFn = func() ([]string, error) { return names, err }
	t.Cleanup(func() { appletNamesFn = orig })
}

func run(t *testing.T, args []string) (string, string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestConfigName(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"cat":         "CAT",
		"add-shell":   "ADD_SHELL",
		"kbd_mode":    "KBD_MODE",
		"rpm2cpio":    "RPM2CPIO",
		"[":           "_",
		"valid-shell": "VALID_SHELL",
	}
	for in, want := range cases {
		if got := configName(in); got != want {
			t.Errorf("configName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseList(t *testing.T) {
	t.Parallel()
	in := "  cat - Concatenate files\n  ls - List files\n\n  wc - Count\n"
	got := parseList(bytes.NewBufferString(in))
	want := []string{"cat", "ls", "wc"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("parseList = %v, want %v", got, want)
	}
}

func TestRunConfig(t *testing.T) {
	withApplets(t, []string{"ls", "cat", "kbd_mode"}, nil)
	out, _, err := run(t, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"CONFIG_NUM_APPLETS=3", "CONFIG_CAT=y", "CONFIG_KBD_MODE=y", "CONFIG_LS=y"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	// Sorted: cat before kbd_mode before ls.
	if strings.Index(out, "CONFIG_CAT=y") > strings.Index(out, "CONFIG_LS=y") {
		t.Errorf("applets not sorted:\n%s", out)
	}
}

func TestRunNames(t *testing.T) {
	withApplets(t, []string{"ls", "cat"}, nil)
	out, _, err := run(t, []string{"--names"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "cat\nls\n" {
		t.Errorf("names output = %q", out)
	}
}

func TestRunUnexpectedArg(t *testing.T) {
	withApplets(t, []string{"ls"}, nil)
	if _, _, err := run(t, []string{"extra"}); err == nil {
		t.Error("expected error for unexpected argument")
	}
}

func TestRunListError(t *testing.T) {
	withApplets(t, nil, context.DeadlineExceeded)
	if _, _, err := run(t, nil); err == nil {
		t.Error("expected error when applet list cannot be read")
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, []string{"--help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Usage: bbconfig") {
		t.Errorf("help missing usage line:\n%s", out)
	}
}
