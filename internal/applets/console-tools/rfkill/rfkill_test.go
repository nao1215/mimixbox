package rfkill

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	mk := func(name string, attrs map[string]string) {
		d := filepath.Join(dir, name)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		for k, v := range attrs {
			if err := os.WriteFile(filepath.Join(d, k), []byte(v+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Out of index order on disk to exercise sorting.
	mk("rfkill1", map[string]string{"name": "hci0", "type": "bluetooth", "soft": "0", "hard": "0"})
	mk("rfkill0", map[string]string{"name": "phy0", "type": "wlan", "soft": "1", "hard": "0"})
	mk("notrfkill", map[string]string{"name": "x"})
	orig := sysClassRfkill
	sysClassRfkill = dir
	t.Cleanup(func() { sysClassRfkill = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestList(t *testing.T) {
	fixture(t)
	out, err := run(t, "list")
	if err != nil {
		t.Fatal(err)
	}
	// Sorted by index: 0 (wlan, soft yes) then 1 (bluetooth).
	if !strings.Contains(out, "0: phy0: wlan") || !strings.Contains(out, "1: hci0: bluetooth") {
		t.Errorf("list output:\n%s", out)
	}
	if strings.Index(out, "phy0") > strings.Index(out, "hci0") {
		t.Errorf("devices not sorted by index:\n%s", out)
	}
	if !strings.Contains(out, "Soft blocked: yes") {
		t.Errorf("wlan should be soft blocked:\n%s", out)
	}
}

func TestDefaultsToList(t *testing.T) {
	fixture(t)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "phy0") {
		t.Errorf("no command should list devices:\n%s", out)
	}
}

func TestBlockUnblock(t *testing.T) {
	fixture(t)
	type call struct {
		index int
		block bool
	}
	got := &call{index: -1}
	orig := blockFn
	blockFn = func(index int, block bool) error { *got = call{index, block}; return nil }
	defer func() { blockFn = orig }()

	if _, err := run(t, "block", "0"); err != nil {
		t.Fatal(err)
	}
	if *got != (call{0, true}) {
		t.Errorf("block call = %+v", *got)
	}
	if _, err := run(t, "unblock", "1"); err != nil {
		t.Fatal(err)
	}
	if *got != (call{1, false}) {
		t.Errorf("unblock call = %+v", *got)
	}
}

func TestErrors(t *testing.T) {
	fixture(t)
	if _, err := run(t, "bogus"); err == nil {
		t.Errorf("an unknown command should fail")
	}
	if _, err := run(t, "block"); err == nil {
		t.Errorf("block without an index should fail")
	}
	if _, err := run(t, "block", "x"); err == nil {
		t.Errorf("a non-numeric index should fail")
	}
}
