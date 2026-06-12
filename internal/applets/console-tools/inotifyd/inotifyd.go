// Package inotifyd implements the inotifyd applet: watch files for inotify
// events and run a handler program for each event.
package inotifyd

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the inotifyd applet.
type Command struct{}

// New returns an inotifyd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "inotifyd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a handler on file inotify events" }

// maskBits maps inotify event flags to the BusyBox inotifyd event letters, in
// bit order.
var maskBits = []struct {
	flag uint32
	ch   byte
}{
	{unix.IN_ACCESS, 'a'}, {unix.IN_MODIFY, 'c'}, {unix.IN_ATTRIB, 'e'},
	{unix.IN_CLOSE_WRITE, 'w'}, {unix.IN_CLOSE_NOWRITE, '0'}, {unix.IN_OPEN, 'r'},
	{unix.IN_MOVED_FROM, 'm'}, {unix.IN_MOVED_TO, 'y'}, {unix.IN_CREATE, 'n'},
	{unix.IN_DELETE, 'd'}, {unix.IN_DELETE_SELF, 'D'}, {unix.IN_MOVE_SELF, 'M'},
	{unix.IN_UNMOUNT, 'u'},
}

// letterFlags is the reverse mapping, used to parse a FILE:mask spec.
var letterFlags = func() map[byte]uint32 {
	m := map[byte]uint32{}
	for _, b := range maskBits {
		m[b.ch] = b.flag
	}
	return m
}()

// event is one inotify event delivered for a watched path.
type event struct {
	mask uint32
	path string // the watched path
	name string // the file within a watched directory, if any
}

// source yields inotify events for the watched paths.
type source interface {
	Recv() (event, error)
	Close() error
}

// watch is a parsed FILE[:mask] specification.
type watch struct {
	path string
	mask uint32
}

// Injected so the watcher and the handler are testable without inotify.
var (
	dialFn    = openInotify
	handlerFn = func(stdio command.IO, prog, actions, path, name string) {
		argv := []string{actions, path}
		if name != "" { // the NAME argument is omitted for an event on the watched file itself
			argv = append(argv, name)
		}
		cmd := exec.Command(prog, argv...) //nolint:gosec // running the handler is the point
		cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
		_ = cmd.Run()
	}
)

// Run executes inotifyd.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "PROG FILE[:MASK]...", stdio.Err).WithHelp(command.Help{
		Description: "Watch each FILE for inotify events and run PROG for each one, as 'PROG ACTIONS " +
			"FILE [NAME]', where ACTIONS is the letters of the events that occurred (a access, c " +
			"modify, e attrib, w close-write, r open, n create, d delete, y moved-to, …). A FILE may be " +
			"suffixed with ':MASK' (those letters) to watch only those events. Runs until interrupted.",
		Examples: []command.Example{
			{Command: "inotifyd ./handler /etc:n /var/log/app.log:w", Explain: "Watch a dir and a file."},
		},
		ExitStatus: "0  the watcher stopped cleanly.\n1  a usage error or the watch could not be set up.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		return command.Failuref("a handler program and at least one file are required")
	}
	prog := rest[0]
	watches, err := parseWatches(rest[1:])
	if err != nil {
		return command.Failuref("%v", err)
	}

	src, err := dialFn(watches)
	if err != nil {
		return command.Failuref("cannot watch: %v", err)
	}
	defer func() { _ = src.Close() }()

	go func() {
		<-ctx.Done()
		_ = src.Close()
	}()

	for {
		ev, err := src.Recv()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, io.EOF) {
				return nil // cancelled, or the source ended cleanly
			}
			return command.Failuref("watch error: %v", err)
		}
		handlerFn(stdio, prog, maskToLetters(ev.mask), ev.path, ev.name)
	}
}

// parseWatches parses the FILE[:MASK] specifications.
func parseWatches(specs []string) ([]watch, error) {
	watches := make([]watch, 0, len(specs))
	for _, s := range specs {
		path, letters, hasMask := strings.Cut(s, ":")
		w := watch{path: path, mask: unix.IN_ALL_EVENTS}
		if hasMask {
			w.mask = 0
			for i := 0; i < len(letters); i++ {
				flag, ok := letterFlags[letters[i]]
				if !ok {
					return nil, command.Failuref("unknown event letter: %q", string(letters[i]))
				}
				w.mask |= flag
			}
		}
		watches = append(watches, w)
	}
	return watches, nil
}

// maskToLetters renders an event mask as its BusyBox event letters.
func maskToLetters(mask uint32) string {
	var b strings.Builder
	for _, mb := range maskBits {
		if mask&mb.flag != 0 {
			b.WriteByte(mb.ch)
		}
	}
	return b.String()
}

// inotifySource reads events from a real inotify file descriptor.
type inotifySource struct {
	fd        int
	wdPath    map[int32]string
	buf       []byte
	queue     []event
	closeOnce sync.Once
	closeErr  error
}

// openInotify sets up an inotify watch for each path.
func openInotify(watches []watch) (source, error) {
	fd, err := unix.InotifyInit1(0)
	if err != nil {
		return nil, err
	}
	s := &inotifySource{fd: fd, wdPath: map[int32]string{}, buf: make([]byte, 8192)}
	for _, w := range watches {
		wd, err := unix.InotifyAddWatch(fd, w.path, w.mask)
		if err != nil {
			_ = unix.Close(fd)
			return nil, err
		}
		s.wdPath[int32(wd)] = w.path
	}
	return s, nil
}

// Recv returns the next inotify event, blocking until one arrives.
func (s *inotifySource) Recv() (event, error) {
	for len(s.queue) == 0 {
		n, err := unix.Read(s.fd, s.buf)
		if err != nil {
			return event{}, err
		}
		s.parse(s.buf[:n])
	}
	ev := s.queue[0]
	s.queue = s.queue[1:]
	return ev, nil
}

// parse splits a read buffer into the individual inotify_event records.
func (s *inotifySource) parse(data []byte) {
	for off := 0; off+16 <= len(data); {
		wd := int32(binary.LittleEndian.Uint32(data[off:]))
		mask := binary.LittleEndian.Uint32(data[off+4:])
		nameLen := int(binary.LittleEndian.Uint32(data[off+12:]))
		name := ""
		if nameLen > 0 && off+16+nameLen <= len(data) {
			name = string(bytes.TrimRight(data[off+16:off+16+nameLen], "\x00"))
		}
		s.queue = append(s.queue, event{mask: mask, path: s.wdPath[wd], name: name})
		off += 16 + nameLen
	}
}

// Close closes the inotify descriptor, idempotently (it is closed from both the
// cancellation goroutine and the deferred cleanup).
func (s *inotifySource) Close() error {
	s.closeOnce.Do(func() { s.closeErr = unix.Close(s.fd) })
	return s.closeErr
}
