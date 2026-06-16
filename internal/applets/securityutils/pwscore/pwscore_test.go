package pwscore

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestCommonPasswordScoresZero(t *testing.T) {
	t.Parallel()
	score, reasons := Score("password")
	if score != 0 {
		t.Errorf("score = %d, want 0", score)
	}
	if len(reasons) == 0 || !strings.Contains(reasons[0], "common") {
		t.Errorf("reasons = %v", reasons)
	}
}

func TestStrongPassword(t *testing.T) {
	t.Parallel()
	score, _ := Score("Tr0ub4dour&3xtraLong!")
	if score < 80 {
		t.Errorf("score = %d, want >= 80 (strong)", score)
	}
}

func TestShortSimpleIsWeak(t *testing.T) {
	t.Parallel()
	score, _ := Score("aaa")
	if score >= 50 {
		t.Errorf("score = %d, want a low score", score)
	}
}

func TestClassify(t *testing.T) {
	t.Parallel()
	n, names := classify("aB1!")
	if n != 4 {
		t.Errorf("classes = %d, want 4; names=%v", n, names)
	}
	n, _ = classify("abcdef")
	if n != 1 {
		t.Errorf("classes = %d, want 1", n)
	}
}

func TestRating(t *testing.T) {
	t.Parallel()
	cases := map[int]string{90: "strong", 60: "fair", 30: "weak", 10: "very weak"}
	for score, want := range cases {
		if got := rating(score); got != want {
			t.Errorf("rating(%d) = %q, want %q", score, got, want)
		}
	}
}

func TestRunFromArg(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "Sup3rSecret!Phrase")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Score:") {
		t.Errorf("out = %q", out)
	}
}

func TestRunFromStdin(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "hunter2\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Score:") {
		t.Errorf("out = %q", out)
	}
}

func TestRunEmpty(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no password provided") {
		t.Errorf("err = %v", err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "pwscore" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Examples:") {
		t.Errorf("--help missing Examples section:\n%s", got)
	}
	if !strings.Contains(got, "Exit status:") {
		t.Errorf("--help missing Exit status section:\n%s", got)
	}
}
