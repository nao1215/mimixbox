package ascii

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func TestPrintsFullTable(t *testing.T) {
	lines := strings.Split(strings.TrimRight(run(t), "\n"), "\n")
	if len(lines) != 128 {
		t.Fatalf("printed %d lines, want 128", len(lines))
	}
}

func TestSpecificEntries(t *testing.T) {
	lines := strings.Split(strings.TrimRight(run(t), "\n"), "\n")
	cases := map[int]string{
		0:   "NUL",
		10:  "LF",
		27:  "ESC",
		32:  "SP",
		65:  "A",
		97:  "a",
		127: "DEL",
	}
	for code, want := range cases {
		line := lines[code]
		if !strings.HasSuffix(line, "  "+want) {
			t.Errorf("line %d = %q, want it to end with %q", code, line, want)
		}
		if !strings.Contains(line, fmtHex(code)) {
			t.Errorf("line %d = %q, missing hex %s", code, line, fmtHex(code))
		}
	}
}

func fmtHex(code int) string {
	const digits = "0123456789ABCDEF"
	return "0x" + string([]byte{digits[code>>4], digits[code&0xF]})
}

func TestReprHelper(t *testing.T) {
	t.Parallel()
	if repr(7) != "BEL" || repr(126) != "~" || repr(127) != "DEL" || repr(32) != "SP" {
		t.Errorf("repr mapping wrong")
	}
}
