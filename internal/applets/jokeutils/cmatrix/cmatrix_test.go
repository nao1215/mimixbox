package cmatrix

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// fixedGlyph lights every cell with '#', so trails are easy to count.
func fixedGlyph(_, _ int) rune { return '#' }

func TestRenderFrameDrawsTrail(t *testing.T) {
	t.Parallel()
	// One column, head at row 5, height 10: rows 0..5 with a trailLen trail.
	frame := RenderFrame(1, 10, []int{5}, fixedGlyph)
	// Each of the 10 rows is newline-terminated, so Split yields a trailing "".
	lines := strings.Split(frame, "\n")
	lines = lines[:len(lines)-1]
	if len(lines) != 10 {
		t.Fatalf("got %d rows, want 10", len(lines))
	}
	lit := 0
	for _, l := range lines {
		if strings.Contains(l, "#") {
			lit++
		}
	}
	if lit != trailLen {
		t.Errorf("lit rows = %d, want %d", lit, trailLen)
	}
}

func TestRenderFrameHeadPosition(t *testing.T) {
	t.Parallel()
	frame := RenderFrame(1, 10, []int{5}, fixedGlyph)
	lines := strings.Split(frame, "\n")
	if lines[5] != "#" {
		t.Errorf("head row = %q, want #", lines[5])
	}
	if lines[6] != "" {
		t.Errorf("row below head should be blank, got %q", lines[6])
	}
}

func TestAdvanceMovesDown(t *testing.T) {
	t.Parallel()
	got := advance([]int{0, 3}, 10, func() int { return 0 })
	if got[0] != 1 || got[1] != 4 {
		t.Errorf("advance = %v, want [1 4]", got)
	}
}

func TestAdvanceWraps(t *testing.T) {
	t.Parallel()
	// A head well past the bottom restarts at the value next() returns.
	got := advance([]int{100}, 10, func() int { return -2 })
	if got[0] != -2 {
		t.Errorf("advance wrap = %v, want [-2]", got)
	}
}

func TestRandomGlyphInAlphabet(t *testing.T) {
	t.Parallel()
	for i := 0; i < 50; i++ {
		if !strings.ContainsRune(glyphs, randomGlyph()) {
			t.Fatal("randomGlyph returned a rune outside the alphabet")
		}
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
	if c.Name() != "cmatrix" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestHelpSections verifies that --help renders both the Examples and the
// Exit status sections supplied through WithHelp.
func TestHelpSections(t *testing.T) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q section:\n%s", want, out.String())
		}
	}
}
