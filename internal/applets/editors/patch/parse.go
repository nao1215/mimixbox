package patch

import (
	"fmt"
	"strconv"
	"strings"
)

// hunk is one @@ block of a unified diff.
type hunk struct {
	oldStart int
	oldLen   int
	newStart int
	newLen   int
	lines    []string // each begins with ' ', '-' or '+'
}

// filePatch is the set of hunks for one file, with its --- and +++ names.
type filePatch struct {
	oldName string
	newName string
	hunks   []hunk
}

// parseUnified parses unified-diff text into per-file patches.
func parseUnified(text string) ([]filePatch, error) {
	lines := strings.Split(text, "\n")
	var patches []filePatch
	var cur *filePatch
	i := 0
	for i < len(lines) {
		line := lines[i]
		switch {
		case strings.HasPrefix(line, "--- "):
			if i+1 >= len(lines) || !strings.HasPrefix(lines[i+1], "+++ ") {
				return nil, fmt.Errorf("malformed header near line %d", i+1)
			}
			patches = append(patches, filePatch{
				oldName: strings.TrimSpace(strings.TrimPrefix(line, "--- ")),
				newName: strings.TrimSpace(strings.TrimPrefix(lines[i+1], "+++ ")),
			})
			cur = &patches[len(patches)-1]
			i += 2
		case strings.HasPrefix(line, "@@"):
			if cur == nil {
				return nil, fmt.Errorf("hunk before file header")
			}
			h, next, err := parseHunk(lines, i)
			if err != nil {
				return nil, err
			}
			cur.hunks = append(cur.hunks, h)
			i = next
		default:
			i++
		}
	}
	return patches, nil
}

// parseHunk parses one @@ header and its body starting at index start.
func parseHunk(lines []string, start int) (hunk, int, error) {
	var h hunk
	header := lines[start]
	os1, ol, ns, nl, err := parseHunkHeader(header)
	if err != nil {
		return h, start, err
	}
	h.oldStart, h.oldLen, h.newStart, h.newLen = os1, ol, ns, nl

	i := start + 1
	for i < len(lines) {
		l := lines[i]
		if l == "" && i == len(lines)-1 {
			break
		}
		if strings.HasPrefix(l, "@@") || strings.HasPrefix(l, "--- ") {
			break
		}
		if len(l) == 0 {
			// A truly empty line counts as a context line containing "".
			h.lines = append(h.lines, " ")
			i++
			continue
		}
		switch l[0] {
		case ' ', '-', '+':
			h.lines = append(h.lines, l)
			i++
		case '\\':
			// "\ No newline at end of file" - ignore.
			i++
		default:
			// End of hunk body.
			return h, i, nil
		}
	}
	return h, i, nil
}

// parseHunkHeader parses "@@ -l,s +l,s @@".
func parseHunkHeader(s string) (oldStart, oldLen, newStart, newLen int, err error) {
	fields := strings.Fields(s)
	if len(fields) < 3 || fields[0] != "@@" {
		return 0, 0, 0, 0, fmt.Errorf("bad hunk header: %q", s)
	}
	oldStart, oldLen, err = parseRange(strings.TrimPrefix(fields[1], "-"))
	if err != nil {
		return
	}
	newStart, newLen, err = parseRange(strings.TrimPrefix(fields[2], "+"))
	return
}

// parseRange parses "start,len" or "start" (len defaults to 1).
func parseRange(s string) (start, length int, err error) {
	length = 1
	if idx := strings.IndexByte(s, ','); idx >= 0 {
		length, err = strconv.Atoi(s[idx+1:])
		if err != nil {
			return
		}
		s = s[:idx]
	}
	start, err = strconv.Atoi(s)
	return
}

// reverseHunks swaps the add/remove sense of every hunk so a patch can be
// un-applied.
func reverseHunks(fp *filePatch) {
	for hi := range fp.hunks {
		h := &fp.hunks[hi]
		h.oldStart, h.newStart = h.newStart, h.oldStart
		h.oldLen, h.newLen = h.newLen, h.oldLen
		for li, l := range h.lines {
			if l == "" {
				continue
			}
			switch l[0] {
			case '-':
				h.lines[li] = "+" + l[1:]
			case '+':
				h.lines[li] = "-" + l[1:]
			}
		}
	}
}

// applyHunks applies every hunk to orig and returns the patched lines.
func applyHunks(orig []string, hunks []hunk) ([]string, error) {
	var out []string
	pos := 0 // 0-based index into orig
	for _, h := range hunks {
		target := h.oldStart - 1
		if target < 0 {
			target = 0
		}
		if target > len(orig) {
			return nil, fmt.Errorf("hunk start %d beyond end of file", h.oldStart)
		}
		// Copy unchanged lines before the hunk.
		out = append(out, orig[pos:target]...)
		pos = target

		for _, l := range h.lines {
			tag, text := l[0], l[1:]
			switch tag {
			case ' ':
				if pos >= len(orig) || orig[pos] != text {
					return nil, fmt.Errorf("context mismatch at line %d", pos+1)
				}
				out = append(out, orig[pos])
				pos++
			case '-':
				if pos >= len(orig) || orig[pos] != text {
					return nil, fmt.Errorf("delete mismatch at line %d", pos+1)
				}
				pos++
			case '+':
				out = append(out, text)
			}
		}
	}
	// Copy whatever remains after the last hunk.
	out = append(out, orig[pos:]...)
	return out, nil
}
