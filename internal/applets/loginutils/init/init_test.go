package initapplet

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type ranEntry struct {
	process string
	wait    bool
}

func setup(t *testing.T, inittab string) *[]ranEntry {
	t.Helper()
	p := filepath.Join(t.TempDir(), "inittab")
	if err := os.WriteFile(p, []byte(inittab), 0o644); err != nil {
		t.Fatal(err)
	}
	var ran []ranEntry
	oi, orf := inittabPath, runFn
	inittabPath = p
	runFn = func(_ command.IO, process string, wait bool) error {
		ran = append(ran, ranEntry{process, wait})
		return nil
	}
	t.Cleanup(func() { inittabPath, runFn = oi, orf })
	return &ran
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestRunsOneShotActionsInOrder(t *testing.T) {
	ran := setup(t, "# comment\n\nsi::sysinit:mount -a\nl0::wait:run-stuff\nx::once:spawn-bg\ntty1::respawn:getty tty1\n")
	if err := run(t); err != nil {
		t.Fatal(err)
	}
	want := []ranEntry{
		{"mount -a", true},   // sysinit: waited
		{"run-stuff", true},  // wait: waited
		{"spawn-bg", false},  // once: backgrounded
	}
	if len(*ran) != len(want) {
		t.Fatalf("ran %v, want %v", *ran, want)
	}
	for i, w := range want {
		if (*ran)[i] != w {
			t.Errorf("entry %d = %+v, want %+v", i, (*ran)[i], w)
		}
	}
}

func TestParseInittab(t *testing.T) {
	t.Parallel()
	entries := parseInittab("# c\n\nid:rl:action:the:process\nbad-line\nx::once:cmd\n")
	if len(entries) != 2 {
		t.Fatalf("parsed %d, want 2", len(entries))
	}
	// id:runlevels:action:process — SplitN(4) keeps colons in the process verbatim.
	if entries[0].action != "action" || entries[0].process != "the:process" {
		t.Errorf("entry 0 = %+v", entries[0])
	}
}

func TestAlias(t *testing.T) {
	if New().Name() != "init" || NewLinuxrc().Name() != "linuxrc" {
		t.Errorf("alias names wrong")
	}
}

func TestMissingInittab(t *testing.T) {
	setup(t, "")
	if err := run(t, "-t", "/no/such/inittab"); err == nil {
		t.Errorf("a missing inittab should fail")
	}
}
