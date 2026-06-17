package mbsh

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/testutil/fakecmd"
)

func parseLine(t *testing.T, input string) commandList {
	t.Helper()
	toks, err := tokenize(input, 0)
	if err != nil {
		t.Fatalf("tokenize(%q) error = %v", input, err)
	}
	list, err := parse(toks)
	if err != nil {
		t.Fatalf("parse(%q) error = %v", input, err)
	}
	return list
}

func TestParseStructure(t *testing.T) {
	t.Run("sequence", func(t *testing.T) {
		list := parseLine(t, "echo a; echo b")
		if len(list.pipelines) != 2 {
			t.Fatalf("pipelines = %d, want 2", len(list.pipelines))
		}
	})
	t.Run("pipeline", func(t *testing.T) {
		list := parseLine(t, "echo a | wc -c")
		if len(list.pipelines) != 1 || len(list.pipelines[0].cmds) != 2 {
			t.Fatalf("pipeline shape = %+v", list.pipelines)
		}
		if got := list.pipelines[0].cmds[1].args; strings.Join(got, " ") != "wc -c" {
			t.Errorf("second command = %v, want [wc -c]", got)
		}
	})
	t.Run("redirections", func(t *testing.T) {
		list := parseLine(t, "cat < in.txt > out.txt")
		redirs := list.pipelines[0].cmds[0].redirs
		if len(redirs) != 2 || redirs[0].op != "<" || redirs[0].file != "in.txt" || redirs[1].op != ">" || redirs[1].file != "out.txt" {
			t.Errorf("redirs = %+v", redirs)
		}
	})
	t.Run("append", func(t *testing.T) {
		list := parseLine(t, "echo x >> log")
		redirs := list.pipelines[0].cmds[0].redirs
		if len(redirs) != 1 || redirs[0].op != ">>" {
			t.Errorf("redirs = %+v, want one >> redirect", redirs)
		}
	})
}

func TestParseErrors(t *testing.T) {
	for _, input := range []string{"| echo", "echo |", "; echo", "echo >", "echo < "} {
		toks, err := tokenize(input, 0)
		if err != nil {
			continue // a tokenizer error is also a rejection
		}
		if _, err := parse(toks); err == nil {
			t.Errorf("parse(%q) should have failed", input)
		}
	}
}

// execLine drives the executor for one input line and returns stdout and the
// list's exit status.
func execLine(t *testing.T, input string) (string, int) {
	t.Helper()
	sh := &shell{}
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	list := parseLine(t, input)
	status := sh.execList(context.Background(), io, list)
	return out.String(), status
}

// requireCmds installs repo-local fakes for the given command names at the
// front of PATH so the executor's exec path runs deterministically without
// depending on host commands.
func requireCmds(t *testing.T, names ...string) {
	t.Helper()
	fakecmd.Use(t, names...)
}

func TestExecSequence(t *testing.T) {
	requireCmds(t, "echo")
	out, _ := execLine(t, "echo one; echo two")
	if out != "one\ntwo\n" {
		t.Errorf("sequence output = %q, want %q", out, "one\ntwo\n")
	}
}

func TestExecPipeline(t *testing.T) {
	requireCmds(t, "printf", "wc")
	out, status := execLine(t, "printf foo | wc -c")
	if strings.TrimSpace(out) != "3" {
		t.Errorf("pipeline output = %q, want 3", out)
	}
	if status != 0 {
		t.Errorf("pipeline status = %d, want 0", status)
	}
}

func TestExecRedirections(t *testing.T) {
	requireCmds(t, "echo", "cat", "wc")
	dir := t.TempDir()
	out := filepath.Join(dir, "out.txt")

	execLine(t, "echo hi > "+out)
	if got, _ := os.ReadFile(out); string(got) != "hi\n" {
		t.Errorf("after > the file = %q, want %q", got, "hi\n")
	}

	execLine(t, "echo more >> "+out)
	if got, _ := os.ReadFile(out); string(got) != "hi\nmore\n" {
		t.Errorf("after >> the file = %q, want %q", got, "hi\nmore\n")
	}

	stdout, _ := execLine(t, "wc -l < "+out)
	if strings.TrimSpace(stdout) != "2" {
		t.Errorf("wc -l < file = %q, want 2", stdout)
	}
}

// TestExecPipelineStatusIsLast documents that a pipeline's status is its last
// command's.
func TestExecPipelineStatusIsLast(t *testing.T) {
	requireCmds(t, "true", "false")
	if _, status := execLine(t, "false | true"); status != 0 {
		t.Errorf("false | true status = %d, want 0 (last command)", status)
	}
	if _, status := execLine(t, "true | false"); status == 0 {
		t.Errorf("true | false status = 0, want non-zero (last command)")
	}
}

func TestExecExitStops(t *testing.T) {
	sh := &shell{}
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	list := parseLine(t, "exit")
	sh.execList(context.Background(), io, list)
	if !sh.stop {
		t.Errorf("exit should set the stop flag")
	}
}
