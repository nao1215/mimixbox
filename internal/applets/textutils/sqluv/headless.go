//
// mimixbox/internal/applets/textutils/sqluv/headless.go
//
// Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sqluv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// runHeadless loads sources, executes opts.execute, and writes the result in
// the requested output format. It also appends the query to the history file.
func runHeadless(stdio command.IO, sources []string, opts options) error {
	if err := validateOutputFormat(opts.output); err != nil {
		return err
	}

	eng, err := loadAll(sources)
	if err != nil {
		return err
	}
	defer func() { _ = eng.close() }()

	rs, err := eng.query(opts.execute, opts.readOnly)
	if err != nil {
		return err
	}

	if err := recordHistory(opts.historyFile, opts.execute); err != nil {
		// History is best-effort; a failure to record it must not fail the query.
		_, _ = fmt.Fprintf(stdio.Err, "sqluv: warning: could not write history: %v\n", err)
	}

	return writeResult(stdio.Out, rs, opts.output)
}

// loadAll builds an engine and loads every source into it.
func loadAll(sources []string) (*engine, error) {
	eng, err := newEngine()
	if err != nil {
		return nil, err
	}
	for _, src := range sources {
		switch classifySource(src) {
		case kindDelimited:
			t, lerr := loadDelimited(src)
			if lerr != nil {
				_ = eng.close()
				return nil, lerr
			}
			if lerr := eng.loadTable(t); lerr != nil {
				_ = eng.close()
				return nil, lerr
			}
		case kindSQLite:
			if lerr := eng.loadSQLiteFile(src); lerr != nil {
				_ = eng.close()
				return nil, lerr
			}
		case kindUnsupported:
			_ = eng.close()
			return nil, validateSource(src)
		}
	}
	return eng, nil
}

// validateOutputFormat ensures the --output value is one this applet supports.
func validateOutputFormat(format string) error {
	switch format {
	case "table", "csv", "tsv", "json":
		return nil
	default:
		return fmt.Errorf("unsupported --output format %q (want table, csv, tsv, or json)", format)
	}
}

// writeResult renders rs to w in the named format.
func writeResult(w io.Writer, rs *resultSet, format string) error {
	switch format {
	case "table":
		return writeTable(w, rs)
	case "csv":
		return writeSeparated(w, rs, ',')
	case "tsv":
		return writeSeparated(w, rs, '\t')
	case "json":
		return writeJSON(w, rs)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

// writeTable renders rs as an aligned ASCII table similar to a SQL shell.
func writeTable(w io.Writer, rs *resultSet) error {
	widths := make([]int, len(rs.columns))
	for i, c := range rs.columns {
		widths[i] = len(c)
	}
	for _, row := range rs.rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var b strings.Builder
	writeRule(&b, widths)
	writeRow(&b, rs.columns, widths)
	writeRule(&b, widths)
	for _, row := range rs.rows {
		writeRow(&b, row, widths)
	}
	writeRule(&b, widths)
	fmt.Fprintf(&b, "(%d row%s)\n", len(rs.rows), plural(len(rs.rows)))
	_, err := io.WriteString(w, b.String())
	return err
}

func writeRule(b *strings.Builder, widths []int) {
	b.WriteByte('+')
	for _, wdt := range widths {
		b.WriteString(strings.Repeat("-", wdt+2))
		b.WriteByte('+')
	}
	b.WriteByte('\n')
}

func writeRow(b *strings.Builder, cells []string, widths []int) {
	b.WriteByte('|')
	for i := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		fmt.Fprintf(b, " %-*s |", widths[i], cell)
	}
	b.WriteByte('\n')
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// writeSeparated renders rs as CSV or TSV (header row plus data rows).
func writeSeparated(w io.Writer, rs *resultSet, comma rune) error {
	cw := csv.NewWriter(w)
	cw.Comma = comma
	if err := cw.Write(rs.columns); err != nil {
		return err
	}
	for _, row := range rs.rows {
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// writeJSON renders rs as a JSON array of objects keyed by column name.
func writeJSON(w io.Writer, rs *resultSet) error {
	records := make([]map[string]string, 0, len(rs.rows))
	for _, row := range rs.rows {
		rec := make(map[string]string, len(rs.columns))
		for i, col := range rs.columns {
			if i < len(row) {
				rec[col] = row[i]
			}
		}
		records = append(records, rec)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}
