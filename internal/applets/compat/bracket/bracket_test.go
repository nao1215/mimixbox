package bracket

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *command.ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return 1
}

func run(c *Command, args ...string) (string, int) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := c.Run(context.Background(), io, args)
	return out.String(), exitCode(err)
}

func TestBracket(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		c    *Command
		args []string
		want int
	}{
		{"true", NewBracket(), []string{"1", "=", "1", "]"}, 0},
		{"false", NewBracket(), []string{"1", "=", "2", "]"}, 1},
		{"missing close", NewBracket(), []string{"1", "=", "1"}, 2},
		{"empty", NewBracket(), nil, 2},
		{"double true", NewDoubleBracket(), []string{"-n", "x", "]]"}, 0},
		{"double missing close", NewDoubleBracket(), []string{"-n", "x"}, 2},
		{"double with single close is malformed", NewDoubleBracket(), []string{"-n", "x", "]"}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, got := run(tt.c, tt.args...); got != tt.want {
				t.Errorf("exit = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBracketHelp(t *testing.T) {
	t.Parallel()
	out, code := run(NewBracket(), "--help")
	if code != 0 || !strings.Contains(out, "Usage: [") {
		t.Errorf("--help out=%q code=%d", out, code)
	}
}
