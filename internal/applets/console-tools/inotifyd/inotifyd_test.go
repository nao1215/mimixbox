package inotifyd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

func TestMaskToLetters(t *testing.T) {
	t.Parallel()
	if got := maskToLetters(unix.IN_CREATE); got != "n" {
		t.Errorf("create = %q, want n", got)
	}
	if got := maskToLetters(unix.IN_CREATE | unix.IN_DELETE); got != "nd" {
		t.Errorf("create|delete = %q, want nd", got)
	}
	if got := maskToLetters(unix.IN_ACCESS | unix.IN_CLOSE_WRITE); got != "aw" {
		t.Errorf("access|close-write = %q, want aw", got)
	}
}

func TestParseWatches(t *testing.T) {
	t.Parallel()
	ws, err := parseWatches([]string{"/a", "/b:nc"})
	if err != nil {
		t.Fatal(err)
	}
	if ws[0].mask != unix.IN_ALL_EVENTS {
		t.Errorf("no mask should watch all events")
	}
	if ws[1].mask != unix.IN_CREATE|unix.IN_MODIFY {
		t.Errorf("mask 'nc' = %#x", ws[1].mask)
	}
	if _, err := parseWatches([]string{"/x:Z"}); err == nil {
		t.Errorf("an unknown letter should error")
	}
}

// queueSource yields the queued events, then reports EOF.
type queueSource struct {
	evs []event
	i   int
}

func (q *queueSource) Recv() (event, error) {
	if q.i >= len(q.evs) {
		return event{}, io.EOF
	}
	e := q.evs[q.i]
	q.i++
	return e, nil
}
func (q *queueSource) Close() error { return nil }

type handlerCall struct{ prog, actions, path, name string }

func TestDispatchesEvents(t *testing.T) {
	var calls []handlerCall
	od, oh := dialFn, handlerFn
	dialFn = func([]watch) (source, error) {
		return &queueSource{evs: []event{
			{mask: unix.IN_CREATE, path: "/etc", name: "new.conf"},
			{mask: unix.IN_DELETE, path: "/etc", name: "old.conf"},
		}}, nil
	}
	handlerFn = func(_ command.IO, prog, actions, path, name string) {
		calls = append(calls, handlerCall{prog, actions, path, name})
	}
	defer func() { dialFn, handlerFn = od, oh }()

	io2 := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io2, []string{"./handler", "/etc"}); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 {
		t.Fatalf("got %d handler calls, want 2: %+v", len(calls), calls)
	}
	if calls[0] != (handlerCall{"./handler", "n", "/etc", "new.conf"}) {
		t.Errorf("first call = %+v", calls[0])
	}
	if calls[1] != (handlerCall{"./handler", "d", "/etc", "old.conf"}) {
		t.Errorf("second call = %+v", calls[1])
	}
}

type blockingSource struct{ closed chan struct{} }

func (b *blockingSource) Recv() (event, error) { <-b.closed; return event{}, errors.New("closed") }
func (b *blockingSource) Close() error {
	select {
	case <-b.closed:
	default:
		close(b.closed)
	}
	return nil
}

func TestStopsOnCancel(t *testing.T) {
	od := dialFn
	dialFn = func([]watch) (source, error) { return &blockingSource{closed: make(chan struct{})}, nil }
	defer func() { dialFn = od }()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		io2 := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		done <- New().Run(ctx, io2, []string{"./h", "/etc"})
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("inotifyd did not stop after cancellation")
	}
}

func TestErrors(t *testing.T) {
	io2 := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io2, []string{"./h"}); err == nil {
		t.Errorf("too few args should fail")
	}
	od := dialFn
	dialFn = func([]watch) (source, error) { return nil, errors.New("no such file") }
	defer func() { dialFn = od }()
	if err := New().Run(context.Background(), io2, []string{"./h", "/etc"}); err == nil {
		t.Errorf("a watch setup failure should fail")
	}
}
