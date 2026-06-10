package readprofile

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withProfile(t *testing.T, words []uint32) {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "profile")
	buf := make([]byte, len(words)*4)
	for i, w := range words {
		binary.LittleEndian.PutUint32(buf[i*4:], w)
	}
	if err := os.WriteFile(f, buf, 0o644); err != nil {
		t.Fatal(err)
	}
	orig := profilePath
	profilePath = f
	t.Cleanup(func() { profilePath = orig })
}

func run(t *testing.T) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, nil)
	return out.String(), err
}

func TestSummary(t *testing.T) {
	withProfile(t, []uint32{4, 0, 5, 0, 3}) // step 4, samples 0+5+0+3 = 8
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if out != "profiling step: 4\ntotal samples: 8\n" {
		t.Errorf("readprofile = %q", out)
	}
}

func TestEmptyBuffer(t *testing.T) {
	withProfile(t, nil)
	if _, err := run(t); err == nil {
		t.Errorf("an empty buffer should fail")
	}
}

func TestMissingFile(t *testing.T) {
	orig := profilePath
	profilePath = "/no/such/profile"
	defer func() { profilePath = orig }()
	if _, err := run(t); err == nil {
		t.Errorf("a missing profile should fail")
	}
}

func TestSummarize(t *testing.T) {
	t.Parallel()
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint32(buf[0:], 2)
	binary.LittleEndian.PutUint32(buf[4:], 10)
	binary.LittleEndian.PutUint32(buf[8:], 20)
	binary.LittleEndian.PutUint32(buf[12:], 30)
	step, total := summarize(buf)
	if step != 2 || total != 60 {
		t.Errorf("summarize = %d, %d; want 2, 60", step, total)
	}
}
