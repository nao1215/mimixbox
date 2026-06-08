package rpm_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/archival/rpm"
	"github.com/nao1215/mimixbox/internal/command"
)

type tagval struct {
	tag   int32
	typ   int32
	data  []byte
	count int32
}

func strTag(tag int32, s string) tagval {
	return tagval{tag: tag, typ: 6, data: append([]byte(s), 0), count: 1}
}

func strArrTag(tag int32, ss []string) tagval {
	var b []byte
	for _, s := range ss {
		b = append(b, []byte(s)...)
		b = append(b, 0)
	}
	return tagval{tag: tag, typ: 8, data: b, count: int32(len(ss))}
}

func int32ArrTag(tag int32, vs []int32) tagval {
	b := make([]byte, 4*len(vs))
	for i, v := range vs {
		binary.BigEndian.PutUint32(b[i*4:], uint32(v))
	}
	return tagval{tag: tag, typ: 4, data: b, count: int32(len(vs))}
}

func buildHeader(tags []tagval) []byte {
	var store, index []byte
	for _, t := range tags {
		off := int32(len(store))
		e := make([]byte, 16)
		binary.BigEndian.PutUint32(e[0:], uint32(t.tag))
		binary.BigEndian.PutUint32(e[4:], uint32(t.typ))
		binary.BigEndian.PutUint32(e[8:], uint32(off))
		binary.BigEndian.PutUint32(e[12:], uint32(t.count))
		index = append(index, e...)
		store = append(store, t.data...)
	}
	intro := make([]byte, 16)
	intro[0], intro[1], intro[2], intro[3] = 0x8e, 0xad, 0xe8, 0x01
	binary.BigEndian.PutUint32(intro[8:], uint32(len(tags)))
	binary.BigEndian.PutUint32(intro[12:], uint32(len(store)))
	return append(append(intro, index...), store...)
}

func buildRPM(t *testing.T, mainTags []tagval) string {
	t.Helper()
	lead := make([]byte, 96)
	lead[0], lead[1], lead[2], lead[3] = 0xed, 0xab, 0xee, 0xdb
	var gzbuf bytes.Buffer
	w := gzip.NewWriter(&gzbuf)
	_, _ = w.Write([]byte("x"))
	_ = w.Close()
	out := append([]byte{}, lead...)
	out = append(out, buildHeader(nil)...)
	out = append(out, buildHeader(mainTags)...)
	out = append(out, gzbuf.Bytes()...)

	dir := t.TempDir()
	p := filepath.Join(dir, "pkg.rpm")
	if err := os.WriteFile(p, out, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := rpm.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func samplePkg(t *testing.T) string {
	return buildRPM(t, []tagval{
		strTag(1000, "hello"),
		strTag(1001, "2.10"),
		strTag(1002, "1.fc40"),
		strTag(1004, "A greeting"),
		strTag(1022, "x86_64"),
		strArrTag(1118, []string{"/usr/bin/", "/etc/"}),
		strArrTag(1117, []string{"hello", "hello.conf"}),
		int32ArrTag(1116, []int32{0, 1}),
	})
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := rpm.New()
	if got := c.Name(); got != "rpm" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestQueryNEVRA(t *testing.T) {
	t.Parallel()
	p := samplePkg(t)
	out, errOut, err := run(t, "-qp", p)
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	if strings.TrimSpace(out) != "hello-2.10-1.fc40.x86_64" {
		t.Errorf("out = %q, want hello-2.10-1.fc40.x86_64", out)
	}
}

func TestQueryInfo(t *testing.T) {
	t.Parallel()
	p := samplePkg(t)
	out, _, err := run(t, "-qpi", p)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	for _, want := range []string{"Name        : hello", "Version     : 2.10", "Architecture: x86_64", "Summary     : A greeting"} {
		if !strings.Contains(out, want) {
			t.Errorf("info missing %q in:\n%s", want, out)
		}
	}
}

func TestQueryList(t *testing.T) {
	t.Parallel()
	p := samplePkg(t)
	out, _, err := run(t, "-qpl", p)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(out, "/usr/bin/hello") || !strings.Contains(out, "/etc/hello.conf") {
		t.Errorf("file list wrong:\n%s", out)
	}
}

func TestRequiresQP(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "-q", samplePkg(t))
	if err == nil {
		t.Error("expected error without -p")
	}
	if !strings.Contains(errOut, "package-file queries") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNoFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "-qp")
	if err == nil {
		t.Error("expected error with no package file")
	}
	if !strings.Contains(errOut, "no package file") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "-qp", filepath.Join(t.TempDir(), "nope.rpm"))
	if err == nil {
		t.Error("expected error for missing file")
	}
	if !strings.Contains(errOut, "rpm:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: rpm") {
		t.Errorf("help = %q", out)
	}
}
