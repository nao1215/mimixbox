package factor

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, in string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(in), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestFactorArgs(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"360": "360: 2 2 2 3 3 5\n",
		"97":  "97: 97\n",
		"1":   "1:\n",
		"2":   "2: 2\n",
		"100": "100: 2 2 5 5\n",
	}
	for in, want := range cases {
		got, err := run(t, "", in)
		if err != nil {
			t.Fatalf("factor %s error = %v", in, err)
		}
		if got != want {
			t.Errorf("factor %s = %q, want %q", in, got, want)
		}
	}
}

func TestFactorStdin(t *testing.T) {
	t.Parallel()
	got, err := run(t, "12 13\n")
	if err != nil {
		t.Fatal(err)
	}
	if got != "12: 2 2 3\n13: 13\n" {
		t.Errorf("factor stdin = %q", got)
	}
}

func TestFactorInvalid(t *testing.T) {
	t.Parallel()
	if _, err := run(t, "", "abc"); err == nil {
		t.Errorf("factor abc should fail")
	}
}
