package script

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// clock is indirected so recorded timing is deterministic under test.
var clock = time.Now

// sleep is indirected so replay does not actually wait under test.
var sleep = time.Sleep

// entry is one record of the "delay bytes" timing stream: how long to wait
// before emitting the next chunk and how many bytes that chunk contains.
type entry struct {
	delay float64
	bytes int
}

// recorder tees writes into a buffer while recording per-write timing.
type recorder struct {
	buf    bytes.Buffer
	timing []entry
	last   time.Time
	mirror io.Writer
}

func (r *recorder) Write(p []byte) (int, error) {
	t := clock()
	delay := 0.0
	if !r.last.IsZero() {
		delay = t.Sub(r.last).Seconds()
	}
	r.last = t
	r.timing = append(r.timing, entry{delay: delay, bytes: len(p)})
	if r.mirror != nil {
		_, _ = r.mirror.Write(p)
	}
	return r.buf.Write(p)
}

// formatTiming serializes the recorded timing into the "delay bytes" record
// format that scriptreplay consumes.
func formatTiming(timing []entry) string {
	var tb strings.Builder
	for _, e := range timing {
		fmt.Fprintf(&tb, "%.6f %d\n", e.delay, e.bytes)
	}
	return tb.String()
}

// readTiming parses a "delay bytes" timing file.
func readTiming(path string) ([]entry, error) {
	f, err := os.Open(path) //nolint:gosec // user-named file
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var out []entry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 2 {
			continue
		}
		delay, _ := strconv.ParseFloat(fields[0], 64)
		n, _ := strconv.Atoi(fields[1])
		out = append(out, entry{delay: delay, bytes: n})
	}
	return out, sc.Err()
}

// transcriptBody strips the leading "Script started" header line from a
// recorded typescript; the timing stream covers the bytes after it.
func transcriptBody(data []byte) []byte {
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return data[i+1:]
	}
	return data
}

// replay writes body to out in the chunk sizes recorded in timing, pausing the
// recorded delay before each chunk.
func replay(out io.Writer, body []byte, timing []entry) {
	pos := 0
	for _, e := range timing {
		sleep(time.Duration(e.delay * float64(time.Second)))
		end := pos + e.bytes
		if end > len(body) {
			end = len(body)
		}
		if pos < end {
			_, _ = out.Write(body[pos:end])
		}
		pos = end
	}
}
