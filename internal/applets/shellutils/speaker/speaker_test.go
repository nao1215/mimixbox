package speaker

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func stub(t *testing.T, available string, runErr error) (calledName *string, calledArgs *[]string) {
	t.Helper()
	origLook, origRun := lookPath, runEngine
	name := ""
	var gotArgs []string
	lookPath = func(n string) (string, error) {
		if n == available {
			return "/usr/bin/" + n, nil
		}
		return "", errors.New("not found")
	}
	runEngine = func(n string, a []string) error {
		name, gotArgs = n, a
		return runErr
	}
	t.Cleanup(func() { lookPath, runEngine = origLook, origRun })
	return &name, &gotArgs
}

func TestSpeaksViaEngine(t *testing.T) {
	name, args := stub(t, "espeak", nil)
	out, _, err := run(t, "hello", "world")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if *name != "espeak" {
		t.Errorf("engine = %q, want espeak", *name)
	}
	if got := *args; got[len(got)-1] != "hello world" {
		t.Errorf("args = %v, want text last", got)
	}
	if !strings.Contains(out, "via espeak") {
		t.Errorf("out = %q", out)
	}
}

func TestLanguageOption(t *testing.T) {
	_, args := stub(t, "espeak", nil)
	if _, _, err := run(t, "-l", "fr", "bonjour"); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	joined := strings.Join(*args, " ")
	if !strings.Contains(joined, "-v fr") {
		t.Errorf("args = %v, want -v fr", *args)
	}
}

func TestPrefersFirstAvailable(t *testing.T) {
	name, _ := stub(t, "spd-say", nil)
	if _, _, err := run(t, "hi"); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if *name != "spd-say" {
		t.Errorf("engine = %q, want spd-say", *name)
	}
}

func TestNoEngine(t *testing.T) {
	stub(t, "none-installed", nil)
	_, _, err := run(t, "hi")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no text-to-speech engine") {
		t.Errorf("err = %v", err)
	}
}

func TestEngineFails(t *testing.T) {
	stub(t, "espeak", errors.New("boom"))
	_, _, err := run(t, "hi")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "espeak failed") {
		t.Errorf("err = %v", err)
	}
}

func TestNoText(t *testing.T) {
	t.Parallel()
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no text to speak") {
		t.Errorf("err = %v", err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "speaker" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q section:\n%s", want, out.String())
		}
	}
}
