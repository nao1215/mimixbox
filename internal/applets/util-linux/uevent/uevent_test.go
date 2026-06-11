package uevent

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// queueSource yields the queued messages, then reports EOF.
type queueSource struct {
	msgs [][]byte
	i    int
}

func (q *queueSource) Recv() ([]byte, error) {
	if q.i >= len(q.msgs) {
		return nil, io.EOF
	}
	m := q.msgs[q.i]
	q.i++
	return m, nil
}
func (q *queueSource) Close() error { return nil }

func withSource(t *testing.T, src ueventSource, dialErr error) {
	t.Helper()
	orig := dialFn
	dialFn = func() (ueventSource, error) {
		if dialErr != nil {
			return nil, dialErr
		}
		return src, nil
	}
	t.Cleanup(func() { dialFn = orig })
}

func TestPrintsEvents(t *testing.T) {
	withSource(t, &queueSource{msgs: [][]byte{
		[]byte("add@/devices/pci0000:00/usb1\x00ACTION=add\x00SUBSYSTEM=usb\x00"),
		[]byte("remove@/devices/virtual/net/eth0\x00ACTION=remove\x00"),
	}}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := &bytes.Buffer{}
	io2 := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(ctx, io2, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "add /devices/pci0000:00/usb1") {
		t.Errorf("missing add event:\n%s", out)
	}
	if !strings.Contains(out.String(), "remove /devices/virtual/net/eth0") {
		t.Errorf("missing remove event:\n%s", out)
	}
}

// blockingSource blocks in Recv until it is closed.
type blockingSource struct{ closed chan struct{} }

func (b *blockingSource) Recv() ([]byte, error) {
	<-b.closed
	return nil, errors.New("closed")
}
func (b *blockingSource) Close() error {
	select {
	case <-b.closed:
	default:
		close(b.closed)
	}
	return nil
}

func TestStopsOnCancel(t *testing.T) {
	withSource(t, &blockingSource{closed: make(chan struct{})}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		io2 := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		done <- New().Run(ctx, io2, nil)
	}()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned %v after cancel", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("uevent did not stop after cancellation")
	}
}

func TestDialFailure(t *testing.T) {
	withSource(t, nil, errors.New("operation not permitted"))
	io2 := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io2, nil); err == nil {
		t.Errorf("a dial failure should fail")
	}
}

func TestParseEvent(t *testing.T) {
	t.Parallel()
	a, d := parseEvent([]byte("change@/devices/foo\x00KEY=val\x00"))
	if a != "change" || d != "/devices/foo" {
		t.Errorf("parseEvent = %q, %q", a, d)
	}
	if a, _ := parseEvent([]byte("malformed-no-at")); a != "" {
		t.Errorf("malformed header should yield empty action")
	}
}
