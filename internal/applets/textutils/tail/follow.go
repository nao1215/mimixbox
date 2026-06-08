package tail

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// followTarget tracks the state of a single file that tail is following: the
// open descriptor, how many bytes have already been emitted, and the file
// identity (for -F rotation detection via os.SameFile).
type followTarget struct {
	path   string
	file   *os.File
	info   os.FileInfo
	offset int64
}

// newFollowTargets opens each path positioned at end of file so that only data
// appended after tail starts is emitted. A file that cannot be opened is kept
// as a pending target (file == nil) when retry is set, so -F/--retry can pick
// it up once it appears; otherwise it is skipped (the initial tail pass has
// already reported the error).
func newFollowTargets(paths []string, retry bool) []followTarget {
	targets := make([]followTarget, 0, len(paths))
	for _, path := range paths {
		f, err := os.Open(path) //nolint:gosec // operating on a user-named file is the point
		if err != nil {
			if retry {
				targets = append(targets, followTarget{path: path})
			}
			continue
		}
		info, err := f.Stat()
		if err != nil {
			_ = f.Close()
			continue
		}
		offset, _ := f.Seek(0, io.SeekEnd)
		targets = append(targets, followTarget{path: path, file: f, info: info, offset: offset})
	}
	return targets
}

// closeAll releases every open descriptor held by the targets.
func closeAll(targets []followTarget) {
	for i := range targets {
		if targets[i].file != nil {
			_ = targets[i].file.Close()
		}
	}
}

// follow polls the targets every interval, emitting newly appended data, until
// the context is cancelled. reopen enables -F semantics (re-open a file that is
// rotated or recreated); showHeader prints "==> FILE <==" banners when output
// switches between files.
func follow(ctx context.Context, stdio command.IO, targets []followTarget, interval float64, reopen, showHeader bool) {
	if len(targets) == 0 {
		return
	}
	last := ""
	ticker := time.NewTicker(time.Duration(interval * float64(time.Second)))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for i := range targets {
				targets[i].poll(stdio, reopen, showHeader, &last)
			}
		}
	}
}

// poll reads any data appended to a single target since the last poll. With
// reopen set it first checks whether the path now refers to a different file
// (rotation) or has reappeared, and switches to it.
func (t *followTarget) poll(stdio command.IO, reopen, showHeader bool, last *string) {
	if reopen {
		t.maybeReopen(stdio)
	}
	if t.file == nil {
		return
	}
	info, err := t.file.Stat()
	if err != nil {
		return
	}
	size := info.Size()
	if size < t.offset {
		// The file was truncated in place; restart from the beginning.
		_, _ = fmt.Fprintf(stdio.Err, "tail: %s: file truncated\n", t.path)
		if _, err := t.file.Seek(0, io.SeekStart); err != nil {
			return
		}
		t.offset = 0
	}
	if size > t.offset {
		t.emit(stdio, size, showHeader, last)
	}
}

// maybeReopen implements -F: if the path no longer resolves to the descriptor
// tail is holding (rotated, replaced, or recreated after deletion), reopen it
// from the start. A still-missing file is left pending for the next poll.
func (t *followTarget) maybeReopen(stdio command.IO) {
	nameInfo, err := os.Stat(t.path)
	if err != nil {
		if t.file != nil {
			_ = t.file.Close()
			t.file = nil
			t.info = nil
		}
		return
	}
	if t.file != nil && t.info != nil && os.SameFile(t.info, nameInfo) {
		return
	}
	f, err := os.Open(t.path) //nolint:gosec // operating on a user-named file is the point
	if err != nil {
		return
	}
	if t.file != nil {
		_ = t.file.Close()
		_, _ = fmt.Fprintf(stdio.Err, "tail: %s: file has been replaced; following new file\n", t.path)
	} else {
		_, _ = fmt.Fprintf(stdio.Err, "tail: %s: file has appeared; following new file\n", t.path)
	}
	t.file = f
	t.info = nameInfo
	t.offset = 0
}

// emit copies the bytes between the last offset and size to stdout, writing a
// header first when output has switched to a different file.
func (t *followTarget) emit(stdio command.IO, size int64, showHeader bool, last *string) {
	if showHeader && *last != t.path {
		writeHeader(stdio.Out, t.path, *last == "")
		*last = t.path
	}
	n, _ := io.CopyN(stdio.Out, t.file, size-t.offset)
	t.offset += n
}
