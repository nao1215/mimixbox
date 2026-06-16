package xargs_test

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/findutils/xargs"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := xargs.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := xargs.New()
	if got := c.Name(); got != "xargs" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestEchoDefault(t *testing.T) {
	t.Parallel()
	// Default command is echo; all items end up on one line.
	out, errOut, err := run(t, "a b c\n", "echo")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	if strings.TrimSpace(out) != "a b c" {
		t.Errorf("out = %q, want 'a b c'", out)
	}
}

func TestMaxArgs(t *testing.T) {
	t.Parallel()
	// -n 1 runs echo once per item, producing one line each.
	out, errOut, err := run(t, "x y z\n", "-n", "1", "echo")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got := strings.Fields(strings.TrimSpace(out))
	if len(got) != 3 {
		t.Errorf("out = %q, want three items on separate invocations", out)
	}
	if strings.Count(strings.TrimRight(out, "\n"), "\n") != 2 {
		t.Errorf("-n 1 should yield 3 lines, got %q", out)
	}
}

func TestReplace(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "world\n", "-I", "{}", "echo", "hello", "{}")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	if strings.TrimSpace(out) != "hello world" {
		t.Errorf("out = %q, want 'hello world'", out)
	}
}

func TestNullDelimiter(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\x00b\x00c\x00", "-0", "echo")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if strings.TrimSpace(out) != "a b c" {
		t.Errorf("out = %q, want 'a b c'", out)
	}
}

func TestCustomDelimiter(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a,b,c", "-d", ",", "echo")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if strings.TrimSpace(out) != "a b c" {
		t.Errorf("out = %q, want 'a b c'", out)
	}
}

func TestNoRunIfEmpty(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "   \n", "-r", "echo", "should-not-run")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "" {
		t.Errorf("out = %q, want empty (command should not run)", out)
	}
}

func TestVerbose(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "hi\n", "-t", "echo")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(errOut, "echo hi") {
		t.Errorf("verbose stderr = %q, want 'echo hi'", errOut)
	}
}

func TestCommandFailure(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "x\n", "this-command-does-not-exist-xyz")
	if err == nil {
		t.Error("expected error when command cannot be run")
	}
	if !strings.Contains(errOut, "xargs:") {
		t.Errorf("stderr = %q, want xargs: prefix", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: xargs") {
		t.Errorf("help = %q", out)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q:\n%s", want, out)
		}
	}
}

func TestMaxLines(t *testing.T) {
	t.Parallel()
	// -L 1 runs echo once per input line, so each line's items stay together
	// and produce one output line each.
	out, errOut, err := run(t, "a b\nc d\ne f\n", "-L", "1", "echo")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got := strings.Split(strings.TrimRight(out, "\n"), "\n")
	want := []string{"a b", "c d", "e f"}
	if len(got) != len(want) {
		t.Fatalf("-L 1 produced %d lines, want %d: %q", len(got), len(want), out)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestMaxLinesTwo(t *testing.T) {
	t.Parallel()
	// -L 2 groups two input lines per invocation: 3 lines -> 2 invocations.
	out, errOut, err := run(t, "a\nb\nc\n", "-L", "2", "echo")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(got) != 2 {
		t.Fatalf("-L 2 produced %d lines, want 2: %q", len(got), out)
	}
	if got[0] != "a b" || got[1] != "c" {
		t.Errorf("got %q, want ['a b','c']", got)
	}
}

func TestMaxChars(t *testing.T) {
	t.Parallel()
	// -s limits the constructed command-line length, splitting a long input
	// into multiple invocations. Each invocation's appended items must fit the
	// budget. Use single-character items so we can reason about lengths.
	const budget = 6 // "x x x" (5) fits; adding " x" (2) -> 7 overflows.
	out, errOut, err := run(t, "1 2 3 4 5 6 7 8\n", "-s", fmt.Sprintf("%d", budget), "echo")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("-s %d should split into multiple invocations, got %q", budget, out)
	}
	// Every produced batch must respect the character budget (sum of item
	// lengths plus one separating space each).
	for _, ln := range lines {
		items := strings.Fields(ln)
		total := 0
		for _, it := range items {
			total += len(it) + 1
		}
		if len(items) > 1 && total > budget {
			t.Errorf("batch %q length %d exceeds budget %d", ln, total, budget)
		}
	}
	// All eight items must still appear exactly once across all invocations.
	all := strings.Fields(strings.Join(lines, " "))
	if len(all) != 8 {
		t.Errorf("expected 8 items across batches, got %d: %q", len(all), all)
	}
}

func TestMaxProcsParallel(t *testing.T) {
	t.Parallel()
	// -P runs invocations concurrently. Regardless of scheduling order, every
	// batch must still run and all outputs must appear. Combine with -n 1 so we
	// get one invocation per item, then collect the unordered set of outputs.
	const n = 20
	var sb strings.Builder
	want := map[string]bool{}
	for i := 0; i < n; i++ {
		tok := fmt.Sprintf("item%02d", i)
		sb.WriteString(tok)
		sb.WriteByte('\n')
		want[tok] = true
	}
	out, errOut, err := run(t, sb.String(), "-P", "4", "-n", "1", "echo")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got := map[string]bool{}
	for _, ln := range strings.Fields(out) {
		got[ln] = true
	}
	if len(got) != n {
		t.Errorf("got %d distinct outputs, want %d: %q", len(got), n, out)
	}
	for tok := range want {
		if !got[tok] {
			t.Errorf("missing output for %q", tok)
		}
	}
}

func TestMaxProcsZeroAsManyAsPossible(t *testing.T) {
	t.Parallel()
	// -P 0 means "run as many as possible"; all batches must still complete and
	// every output must appear.
	out, errOut, err := run(t, "a\nb\nc\nd\n", "-P", "0", "-n", "1", "echo")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got := map[string]bool{}
	for _, ln := range strings.Fields(out) {
		got[ln] = true
	}
	for _, tok := range []string{"a", "b", "c", "d"} {
		if !got[tok] {
			t.Errorf("missing output for %q: %q", tok, out)
		}
	}
}

func TestTrueCommandRunsOnEmptyWithoutR(t *testing.T) {
	t.Parallel()
	// Without -r, GNU xargs runs the command once even with empty input.
	// Use "true" (a real binary) to avoid output; on platforms without it,
	// skip.
	if runtime.GOOS != "linux" {
		t.Skip("relies on /usr/bin/true")
	}
	_, _, err := run(t, "", "true")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
}
