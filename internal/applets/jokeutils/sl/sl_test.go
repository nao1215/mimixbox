package sl_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/jokeutils/sl"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if sl.New() == nil {
		t.Fatal("New() returned nil")
	}
}

func TestName(t *testing.T) {
	t.Parallel()
	if got := sl.New().Name(); got != "sl" {
		t.Errorf("Name() = %q, want %q", got, "sl")
	}
}

func TestSynopsis(t *testing.T) {
	t.Parallel()
	want := "Cure your bad habit of mistyping"
	if got := sl.New().Synopsis(); got != want {
		t.Errorf("Synopsis() = %q, want %q", got, want)
	}
}

func TestTrainFrame(t *testing.T) {
	t.Parallel()
	frame := sl.TrainFrame(0)
	if frame == "" {
		t.Fatal("TrainFrame(0) returned empty art")
	}
	// A recognizable piece of the steam locomotive ASCII art.
	if !strings.Contains(frame, "_D _|") {
		t.Errorf("TrainFrame(0) missing recognizable art; got:\n%s", frame)
	}

	// A positive offset shifts the art to the right with leading spaces.
	shifted := sl.TrainFrame(4)
	if !strings.HasPrefix(shifted, "    ") {
		t.Errorf("TrainFrame(4) should be indented; got first line %q",
			strings.SplitN(shifted, "\n", 2)[0])
	}

	// A negative offset is treated as zero (no panic, same as offset 0).
	if sl.TrainFrame(-5) != frame {
		t.Errorf("TrainFrame(-5) should equal TrainFrame(0)")
	}
}

// TestRunNonTTY verifies that Run in a non-TTY environment (as under tests/CI)
// returns without error and without hanging: termbox cannot initialize a
// terminal, so the animation is skipped gracefully.
func TestRunNonTTY(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	stdio := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}

	done := make(chan error, 1)
	go func() {
		done <- sl.New().Run(context.Background(), stdio, nil)
	}()

	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := sl.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q section:\n%s", want, out.String())
		}
	}
}
