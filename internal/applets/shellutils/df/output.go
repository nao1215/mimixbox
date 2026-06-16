package df

import (
	"fmt"
	"io"
	"strings"
)

// defaultOutputAll is the field list GNU df uses when --output is given with no
// FIELD_LIST argument. It lists every supported field in canonical order.
const defaultOutputAll = "source,fstype,itotal,iused,iavail,ipcent,size,used,avail,pcent,file,target"

// outputColumn describes one --output field: its header label and how to render
// a cell for a filesystem entry (and for the grand-total row).
type outputColumn struct {
	field  string
	header string
	right  bool // right-align numeric columns
	cell   func(e fsEntry, opts options) string
	// total renders the cell for the grand-total row. nil means "-".
	total func(entries []fsEntry, opts options) string
}

// outputColumns is the registry of supported --output fields, keyed by field
// name. The required minimum (source, fstype, size, used, avail, pcent, target)
// is implemented, plus the inode fields for completeness.
var outputColumns = map[string]outputColumn{
	"source": {
		field: "source", header: "Filesystem",
		cell: func(e fsEntry, _ options) string {
			if e.source == "" {
				return "-"
			}
			return e.source
		},
		total: func(_ []fsEntry, _ options) string { return "total" },
	},
	"fstype": {
		field: "fstype", header: "Type",
		cell: func(e fsEntry, _ options) string {
			if e.fstype == "" {
				return "-"
			}
			return e.fstype
		},
	},
	"itotal": {
		field: "itotal", header: "Inodes", right: true,
		cell:  func(e fsEntry, _ options) string { return fmt.Sprintf("%d", e.stat.Files) },
		total: sumInodes(func(u inodeUsage) uint64 { return u.files }),
	},
	"iused": {
		field: "iused", header: "IUsed", right: true,
		cell:  func(e fsEntry, _ options) string { return fmt.Sprintf("%d", computeInodeUsage(e.stat).used) },
		total: sumInodes(func(u inodeUsage) uint64 { return u.used }),
	},
	"iavail": {
		field: "iavail", header: "IFree", right: true,
		cell:  func(e fsEntry, _ options) string { return fmt.Sprintf("%d", computeInodeUsage(e.stat).free) },
		total: sumInodes(func(u inodeUsage) uint64 { return u.free }),
	},
	"ipcent": {
		field: "ipcent", header: "IUse%", right: true,
		cell: func(e fsEntry, _ options) string { return fmt.Sprintf("%d%%", computeInodeUsage(e.stat).usePct) },
		total: func(entries []fsEntry, _ options) string {
			var files, used uint64
			for _, e := range entries {
				u := computeInodeUsage(e.stat)
				files += u.files
				used += u.used
			}
			return fmt.Sprintf("%d%%", percent(used, files))
		},
	},
	"size": {
		field: "size", header: "Size", right: true,
		cell: func(e fsEntry, opts options) string {
			return scaleSize(computeUsage(e.stat).total, opts.human, opts.blockSize)
		},
		total: sumUsage(func(u usage) uint64 { return u.total }),
	},
	"used": {
		field: "used", header: "Used", right: true,
		cell: func(e fsEntry, opts options) string {
			return scaleSize(computeUsage(e.stat).used, opts.human, opts.blockSize)
		},
		total: sumUsage(func(u usage) uint64 { return u.used }),
	},
	"avail": {
		field: "avail", header: "Avail", right: true,
		cell: func(e fsEntry, opts options) string {
			return scaleSize(computeUsage(e.stat).avail, opts.human, opts.blockSize)
		},
		total: sumUsage(func(u usage) uint64 { return u.avail }),
	},
	"pcent": {
		field: "pcent", header: "Use%", right: true,
		cell: func(e fsEntry, _ options) string { return fmt.Sprintf("%d%%", computeUsage(e.stat).usePct) },
		total: func(entries []fsEntry, _ options) string {
			var used, avail uint64
			for _, e := range entries {
				u := computeUsage(e.stat)
				used += u.used
				avail += u.avail
			}
			return fmt.Sprintf("%d%%", percent(used, used+avail))
		},
	},
	"file": {
		field: "file", header: "File",
		cell: func(e fsEntry, _ options) string {
			if e.target == "" {
				return "-"
			}
			return e.target
		},
	},
	"target": {
		field: "target", header: "Mounted on",
		cell: func(e fsEntry, _ options) string {
			if e.target == "" {
				return "-"
			}
			return e.target
		},
	},
}

// sumUsage builds a total renderer that sums one byte-based usage field.
func sumUsage(pick func(u usage) uint64) func([]fsEntry, options) string {
	return func(entries []fsEntry, opts options) string {
		var sum uint64
		for _, e := range entries {
			sum += pick(computeUsage(e.stat))
		}
		return scaleSize(sum, opts.human, opts.blockSize)
	}
}

// sumInodes builds a total renderer that sums one inode usage field.
func sumInodes(pick func(u inodeUsage) uint64) func([]fsEntry, options) string {
	return func(entries []fsEntry, _ options) string {
		var sum uint64
		for _, e := range entries {
			sum += pick(computeInodeUsage(e.stat))
		}
		return fmt.Sprintf("%d", sum)
	}
}

// parseOutput parses a comma-separated FIELD_LIST into an ordered, validated
// list of field names. Unknown fields are an error.
func parseOutput(spec string) ([]string, error) {
	parts := strings.Split(spec, ",")
	cols := make([]string, 0, len(parts))
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name == "" {
			continue
		}
		if _, ok := outputColumns[name]; !ok {
			return nil, fmt.Errorf("option --output: field '%s' unknown", name)
		}
		cols = append(cols, name)
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("option --output: empty field list")
	}
	return cols, nil
}

// renderColumns prints entries using the selected --output columns, in the
// requested order, with a header and an optional grand-total row.
func renderColumns(w io.Writer, entries []fsEntry, opts options) {
	cols := make([]outputColumn, 0, len(opts.output))
	for _, name := range opts.output {
		cols = append(cols, outputColumns[name])
	}

	// Compute column widths from header and all cells (plus total row).
	widths := make([]int, len(cols))
	rows := make([][]string, 0, len(entries)+1)
	for i, c := range cols {
		widths[i] = len(c.header)
	}
	for _, e := range entries {
		row := make([]string, len(cols))
		for i, c := range cols {
			row[i] = c.cell(e, opts)
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
		rows = append(rows, row)
	}
	var totalRow []string
	if opts.total {
		totalRow = make([]string, len(cols))
		for i, c := range cols {
			if c.total != nil {
				totalRow[i] = c.total(entries, opts)
			} else {
				totalRow[i] = "-"
			}
			if len(totalRow[i]) > widths[i] {
				widths[i] = len(totalRow[i])
			}
		}
	}

	writeOutputLine(w, headerCells(cols), cols, widths)
	for _, row := range rows {
		writeOutputLine(w, row, cols, widths)
	}
	if totalRow != nil {
		writeOutputLine(w, totalRow, cols, widths)
	}
}

func headerCells(cols []outputColumn) []string {
	h := make([]string, len(cols))
	for i, c := range cols {
		h[i] = c.header
	}
	return h
}

// writeOutputLine writes one space-separated, padded row. Numeric columns are
// right-aligned; text columns are left-aligned, mirroring GNU df --output.
func writeOutputLine(w io.Writer, cells []string, cols []outputColumn, widths []int) {
	var b strings.Builder
	for i, cell := range cells {
		if i > 0 {
			b.WriteByte(' ')
		}
		if cols[i].right {
			fmt.Fprintf(&b, "%*s", widths[i], cell)
		} else {
			fmt.Fprintf(&b, "%-*s", widths[i], cell)
		}
	}
	_, _ = io.WriteString(w, strings.TrimRight(b.String(), " ")+"\n")
}
