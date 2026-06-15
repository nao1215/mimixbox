package telnetd

import "io"

// Reader wraps an io.Reader and strips telnet IAC command sequences from the
// byte stream, yielding only application data. It is stateful across Read calls
// so a command sequence split over buffer boundaries is handled correctly: a
// trailing partial command is retained until the rest of it arrives.
type Reader struct {
	src     io.Reader
	pending []byte // filtered data bytes not yet delivered
	raw     []byte // trailing raw bytes that may be a partial IAC sequence
}

// NewReader returns a telnet Reader over src.
func NewReader(src io.Reader) *Reader { return &Reader{src: src} }

// Read fills p with IAC-filtered data bytes.
func (r *Reader) Read(p []byte) (int, error) {
	for len(r.pending) == 0 {
		tmp := make([]byte, 4096)
		m, err := r.src.Read(tmp)
		if m > 0 {
			r.raw = append(r.raw, tmp[:m]...)
			data, rest := stripIACKeepPartial(r.raw)
			r.pending = append(r.pending, data...)
			r.raw = rest
		}
		if err != nil {
			if len(r.pending) == 0 {
				return 0, err
			}
			break
		}
	}
	n := copy(p, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}

// stripIACKeepPartial filters IAC sequences from buf like StripIAC, but stops at
// the first incomplete trailing IAC sequence and returns those raw bytes so the
// caller can prepend the rest on the next read.
func stripIACKeepPartial(buf []byte) (data, partial []byte) {
	out := make([]byte, 0, len(buf))
	i := 0
	for i < len(buf) {
		if buf[i] != iac {
			out = append(out, buf[i])
			i++
			continue
		}
		// buf[i] == iac
		if i+1 >= len(buf) {
			return out, buf[i:] // dangling IAC
		}
		cmd := buf[i+1]
		switch {
		case cmd == iac:
			out = append(out, iac)
			i += 2
		case cmd == sb:
			j := i + 2
			for j < len(buf) && buf[j] != se {
				j++
			}
			if j >= len(buf) {
				return out, buf[i:] // SB not yet terminated
			}
			i = j + 1
		case cmd >= 251 && cmd <= 254:
			if i+2 >= len(buf) {
				return out, buf[i:] // option byte not yet arrived
			}
			i += 3
		default:
			i += 2
		}
	}
	return out, nil
}
