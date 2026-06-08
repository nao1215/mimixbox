package diff

import (
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// contextLines is the number of unchanged lines shown around each change in
// unified output.
const contextLines = 3

// hunk groups a consecutive run of deletes and inserts together with the file
// positions they start at.
type hunk struct {
	aPos int   // 0-based count of a-lines before the hunk
	bPos int   // 0-based count of b-lines before the hunk
	dels []int // a indices deleted
	ins  []int // b indices inserted
}

// collectHunks groups the edit script into change hunks, dropping the equal
// runs between them.
func collectHunks(ops []op) []hunk {
	var hunks []hunk
	i := 0
	for i < len(ops) {
		if ops[i].kind == opEqual {
			i++
			continue
		}
		h := hunk{aPos: ops[i].ai, bPos: ops[i].bi}
		for i < len(ops) && ops[i].kind != opEqual {
			if ops[i].kind == opDelete {
				h.dels = append(h.dels, ops[i].ai)
			} else {
				h.ins = append(h.ins, ops[i].bi)
			}
			i++
		}
		hunks = append(hunks, h)
	}
	return hunks
}

// rangeStr formats a 1-based inclusive line range the way normal diff does:
// a single number when start == end, otherwise "start,end".
func rangeStr(start, end int) string {
	if start == end {
		return fmt.Sprintf("%d", start)
	}
	return fmt.Sprintf("%d,%d", start, end)
}

// writeNormal prints the classic (ed-style) diff output.
func writeNormal(stdio command.IO, a, b []string, ops []op) {
	for _, h := range collectHunks(ops) {
		switch {
		case len(h.ins) == 0: // deletion only
			aRange := rangeStr(h.aPos+1, h.aPos+len(h.dels))
			_, _ = fmt.Fprintf(stdio.Out, "%sd%d\n", aRange, h.bPos)
			printLines(stdio, "< ", a, h.dels)
		case len(h.dels) == 0: // insertion only
			bRange := rangeStr(h.bPos+1, h.bPos+len(h.ins))
			_, _ = fmt.Fprintf(stdio.Out, "%da%s\n", h.aPos, bRange)
			printLines(stdio, "> ", b, h.ins)
		default: // change
			aRange := rangeStr(h.aPos+1, h.aPos+len(h.dels))
			bRange := rangeStr(h.bPos+1, h.bPos+len(h.ins))
			_, _ = fmt.Fprintf(stdio.Out, "%sc%s\n", aRange, bRange)
			printLines(stdio, "< ", a, h.dels)
			_, _ = fmt.Fprintln(stdio.Out, "---")
			printLines(stdio, "> ", b, h.ins)
		}
	}
}

// printLines writes each indexed line of src with the given prefix.
func printLines(stdio command.IO, prefix string, src []string, idx []int) {
	for _, i := range idx {
		_, _ = fmt.Fprintf(stdio.Out, "%s%s\n", prefix, src[i])
	}
}

// writeUnified prints unified diff output with file headers and @@ hunks.
func writeUnified(stdio command.IO, name1, name2 string, a, b []string, ops []op) {
	groups := groupUnified(ops)
	if len(groups) == 0 {
		return
	}
	_, _ = fmt.Fprintf(stdio.Out, "--- %s\n", name1)
	_, _ = fmt.Fprintf(stdio.Out, "+++ %s\n", name2)

	for _, g := range groups {
		aStart, aLen, bStart, bLen := unifiedRange(g)
		_, _ = fmt.Fprintf(stdio.Out, "@@ -%s +%s @@\n", countRange(aStart, aLen), countRange(bStart, bLen))
		for _, o := range g {
			switch o.kind {
			case opEqual:
				_, _ = fmt.Fprintf(stdio.Out, " %s\n", a[o.ai])
			case opDelete:
				_, _ = fmt.Fprintf(stdio.Out, "-%s\n", a[o.ai])
			case opInsert:
				_, _ = fmt.Fprintf(stdio.Out, "+%s\n", b[o.bi])
			}
		}
	}
}

// groupUnified splits the edit script into hunks padded with up to `context`
// unchanged lines, merging hunks whose context windows overlap.
func groupUnified(ops []op) [][]op {
	// Indices of changed ops.
	var changed []int
	for i, o := range ops {
		if o.kind != opEqual {
			changed = append(changed, i)
		}
	}
	if len(changed) == 0 {
		return nil
	}

	var groups [][]op
	start := max0(changed[0] - contextLines)
	end := min(len(ops), changed[0]+1+contextLines)
	for _, ci := range changed[1:] {
		if ci-contextLines <= end {
			end = min(len(ops), ci+1+contextLines)
			continue
		}
		groups = append(groups, ops[start:end])
		start = max0(ci - contextLines)
		end = min(len(ops), ci+1+contextLines)
	}
	groups = append(groups, ops[start:end])
	return groups
}

// unifiedRange returns the 1-based start lines and lengths in a and b for a hunk.
func unifiedRange(g []op) (aStart, aLen, bStart, bLen int) {
	aStart, bStart = -1, -1
	for _, o := range g {
		if o.kind != opInsert {
			if aStart < 0 {
				aStart = o.ai + 1
			}
			aLen++
		}
		if o.kind != opDelete {
			if bStart < 0 {
				bStart = o.bi + 1
			}
			bLen++
		}
	}
	if aStart < 0 {
		aStart = 0
	}
	if bStart < 0 {
		bStart = 0
	}
	return aStart, aLen, bStart, bLen
}

// countRange formats the "start,len" portion of a unified @@ header.
func countRange(start, length int) string {
	if length == 1 {
		return fmt.Sprintf("%d", start)
	}
	return fmt.Sprintf("%d,%d", start, length)
}

func max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
