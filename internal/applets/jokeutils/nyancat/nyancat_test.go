package nyancat

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestFrameHasCatRows(t *testing.T) {
	t.Parallel()
	lines := strings.Split(Frame(0), "\n")
	if len(lines) != len(cat) {
		t.Fatalf("got %d rows, want %d", len(lines), len(cat))
	}
	if !strings.Contains(Frame(0), "o.o") {
		t.Errorf("cat face missing in %q", Frame(0))
	}
}

func TestFrameTrailGrows(t *testing.T) {
	t.Parallel()
	short := Frame(2)
	long := Frame(10)
	if strings.Count(long, "=") <= strings.Count(short, "=") {
		t.Errorf("longer trail should have more '=': %d vs %d",
			strings.Count(long, "="), strings.Count(short, "="))
	}
}

func TestFrameNegativeTrail(t *testing.T) {
	t.Parallel()
	if strings.Contains(Frame(-5), "=") {
		t.Errorf("negative trail should produce no rainbow: %q", Frame(-5))
	}
}

func TestRunNoTerminalDegradesGracefully(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(ctx, io, nil); err != nil {
		t.Fatalf("Run error = %v (should be nil without a terminal)", err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "nyancat" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
