package halt

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteWtmpRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wtmp")
	now := time.Unix(1700000000, 123000)

	if err := writeWtmp(path, now); err != nil {
		t.Fatalf("writeWtmp error = %v", err)
	}

	rec, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rec) != utmpRecordSize {
		t.Fatalf("record size = %d, want %d", len(rec), utmpRecordSize)
	}
	if got := binary.LittleEndian.Uint16(rec[0:]); got != runLevel {
		t.Errorf("ut_type = %d, want %d (RUN_LVL)", got, runLevel)
	}
	if got := cstr(rec[44:76]); got != "shutdown" {
		t.Errorf("ut_user = %q, want %q", got, "shutdown")
	}
	if got := cstr(rec[8:40]); got != "~~" {
		t.Errorf("ut_line = %q, want %q", got, "~~")
	}
	if got := binary.LittleEndian.Uint32(rec[340:]); got != uint32(now.Unix()) {
		t.Errorf("ut_tv.tv_sec = %d, want %d", got, now.Unix())
	}
}

func TestWriteWtmpAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wtmp")
	now := time.Unix(1700000000, 0)
	if err := writeWtmp(path, now); err != nil {
		t.Fatal(err)
	}
	if err := writeWtmp(path, now); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 2*utmpRecordSize {
		t.Errorf("size after two writes = %d, want %d", info.Size(), 2*utmpRecordSize)
	}
}

// cstr returns the NUL-terminated string at the start of b.
func cstr(b []byte) string {
	before, _, _ := bytes.Cut(b, []byte{0})
	return string(before)
}
