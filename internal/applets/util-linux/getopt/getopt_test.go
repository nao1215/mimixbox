package getopt

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(args ...string) (string, int) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	code := 0
	var ee *command.ExitError
	if errors.As(err, &ee) {
		code = ee.Code
	} else if err != nil {
		code = 1
	}
	return strings.TrimRight(out.String(), "\n"), code
}

func TestGetopt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
		code int
	}{
		{
			name: "short and long with permutation",
			args: []string{"-o", "ab:c", "--long", "alpha,beta:", "--", "-a", "-b", "val", "pos1", "--alpha"},
			want: " -a -b 'val' --alpha -- 'pos1'",
		},
		{
			name: "optional short arg",
			args: []string{"-o", "a::", "--", "-a", "-afoo"},
			want: " -a '' -a 'foo' --",
		},
		{
			name: "optional long arg",
			args: []string{"-o", "", "--long", "alpha::", "--", "--alpha", "--alpha=x"},
			want: " --alpha '' --alpha 'x' --",
		},
		{
			name: "clustered short flags",
			args: []string{"-o", "abc", "--", "-abc", "file"},
			want: " -a -b -c -- 'file'",
		},
		{
			name: "long with =value",
			args: []string{"-o", "f:", "--long", "file:", "--", "--file=out.txt", "arg"},
			want: " --file 'out.txt' -- 'arg'",
		},
		{
			name: "legacy form is unquoted",
			args: []string{"abc", "-a", "x"},
			want: " -a -- x",
		},
		{
			name: "quote escaping",
			args: []string{"-o", "a:", "--", "-a", "it's"},
			want: ` -a 'it'\''s' --`,
		},
		{
			name: "unknown option fails",
			args: []string{"-o", "a", "--", "-x"},
			code: 1,
		},
		{
			name: "missing required arg fails",
			args: []string{"-o", "a:", "--", "-a"},
			code: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, code := run(tt.args...)
			if code != tt.code {
				t.Errorf("exit = %d, want %d (out=%q)", code, tt.code, got)
			}
			if tt.want != "" && got != tt.want {
				t.Errorf("out = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, code := run("--help")
	if code != 0 || !strings.Contains(out, "Usage: getopt") {
		t.Errorf("--help out=%q code=%d", out, code)
	}
}
