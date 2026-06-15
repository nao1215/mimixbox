package sendmail

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestAppendToMbox(t *testing.T) {
	now = func() time.Time { return time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC) }
	defer func() { now = time.Now }()

	mbox := filepath.Join(t.TempDir(), "inbox.mbox")
	msg := "Subject: hi\r\n\r\nbody line\r\nFrom the start\r\n"
	io := command.IO{In: strings.NewReader(msg), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"-o", mbox, "-f", "me@example.com", "you@example.com"}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(mbox)
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if !strings.HasPrefix(s, "From me@example.com ") {
		t.Errorf("missing mbox From separator:\n%s", s)
	}
	if !strings.Contains(s, "Subject: hi") || !strings.Contains(s, "body line") {
		t.Errorf("message body not appended:\n%s", s)
	}
	if !strings.Contains(s, ">From the start") {
		t.Errorf("From-line in body not escaped:\n%s", s)
	}
}

func TestAppendTwice(t *testing.T) {
	mbox := filepath.Join(t.TempDir(), "inbox.mbox")
	for i := 0; i < 2; i++ {
		io := command.IO{In: strings.NewReader("Subject: m\n\nx\n"), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		if err := New().Run(context.Background(), io, []string{"-o", mbox}); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := os.ReadFile(mbox)
	if n := strings.Count(string(got), "From MAILER-DAEMON "); n != 2 {
		t.Errorf("expected 2 mbox entries, got %d", n)
	}
}

func TestNoMbox(t *testing.T) {
	io := command.IO{In: strings.NewReader("msg"), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Fatal("expected error without -o")
	}
}
