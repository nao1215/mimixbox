//
// mimixbox/internal/applets/textutils/sqluv/loader.go
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
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/nao1215/mimixbox/internal/command"
	"github.com/ulikunitz/xz"
)

// table is a delimited source loaded into memory: an ordered header plus rows.
type table struct {
	name    string
	format  fileFormat
	columns []string
	rows    [][]string
}

// openSource opens a possibly-compressed local file and returns a reader over
// its decompressed bytes along with a cleanup function. The cleanup closes both
// the decompressor (when one was wrapped) and the underlying file.
func openSource(path string) (io.Reader, func() error, error) {
	f, err := os.Open(path) //nolint:gosec // path comes from a user-supplied operand by design
	if err != nil {
		return nil, nil, err
	}

	_, comp := splitCompression(path)
	switch comp {
	case compNone:
		return f, f.Close, nil
	case compGzip:
		gr, err := gzip.NewReader(f)
		if err != nil {
			_ = f.Close()
			return nil, nil, fmt.Errorf("gzip: %w", err)
		}
		return gr, func() error { _ = gr.Close(); return f.Close() }, nil
	case compBzip2:
		// compress/bzip2 has no Close; wrap the file's Close only.
		return bzip2.NewReader(f), f.Close, nil
	case compXz:
		xr, err := xz.NewReader(bufio.NewReader(f))
		if err != nil {
			_ = f.Close()
			return nil, nil, fmt.Errorf("xz: %w", err)
		}
		return xr, f.Close, nil
	case compZstd:
		zr, err := zstd.NewReader(f)
		if err != nil {
			_ = f.Close()
			return nil, nil, fmt.Errorf("zstd: %w", err)
		}
		return zr, func() error { zr.Close(); return f.Close() }, nil
	default:
		_ = f.Close()
		return nil, nil, fmt.Errorf("unknown compression for %q", path)
	}
}

// loadDelimited reads a delimited source from path and returns it as a table.
// The table name is derived from the file stem and the format from the
// extension (after stripping any compression suffix).
func loadDelimited(path string) (*table, error) {
	base, _ := splitCompression(path)
	format := detectFormatByName(base)
	if format == formatUnknown {
		return nil, fmt.Errorf("cannot detect delimited format from %q (expected .csv, .tsv, or .ltsv)", path)
	}

	r, closeFn, err := openSource(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = closeFn() }()

	t := &table{name: tableNameFor(path), format: format}
	switch format {
	case formatCSV, formatTSV:
		if err := readSeparated(r, t, format); err != nil {
			return nil, err
		}
	case formatLTSV:
		if err := readLTSV(r, t); err != nil {
			return nil, err
		}
	}
	if len(t.columns) == 0 {
		return nil, fmt.Errorf("%q has no columns", path)
	}
	return t, nil
}

// readSeparated parses CSV or TSV content into t, treating the first record as
// the header row.
func readSeparated(r io.Reader, t *table, format fileFormat) error {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1 // tolerate ragged rows; short rows are padded.
	if format == formatTSV {
		cr.Comma = '\t'
	}
	records, err := cr.ReadAll()
	if err != nil {
		return fmt.Errorf("parse %s: %w", format, err)
	}
	if len(records) == 0 {
		return nil
	}
	t.columns = dedupeColumns(records[0])
	width := len(t.columns)
	for _, rec := range records[1:] {
		t.rows = append(t.rows, padRow(rec, width))
	}
	return nil
}

// readLTSV parses LTSV content into t. Each line is a set of "label:value"
// fields; the union of labels (in first-seen order) becomes the columns.
func readLTSV(r io.Reader, t *table) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), command.MaxLineSize)

	colIndex := map[string]int{}
	var records []map[string]string
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if line == "" {
			continue
		}
		rec := map[string]string{}
		for _, field := range strings.Split(line, "\t") {
			label, value, ok := strings.Cut(field, ":")
			if !ok {
				continue
			}
			if _, seen := colIndex[label]; !seen {
				colIndex[label] = len(t.columns)
				t.columns = append(t.columns, label)
			}
			rec[label] = value
		}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("parse ltsv: %w", err)
	}
	for _, rec := range records {
		row := make([]string, len(t.columns))
		for label, idx := range colIndex {
			row[idx] = rec[label]
		}
		t.rows = append(t.rows, row)
	}
	return nil
}

// dedupeColumns ensures column names are unique and non-empty so they can be
// used as SQLite column identifiers.
func dedupeColumns(cols []string) []string {
	seen := map[string]int{}
	out := make([]string, len(cols))
	for i, c := range cols {
		name := strings.TrimSpace(c)
		if name == "" {
			name = fmt.Sprintf("col%d", i+1)
		}
		if n, dup := seen[name]; dup {
			seen[name] = n + 1
			name = fmt.Sprintf("%s_%d", name, n+1)
		} else {
			seen[name] = 1
		}
		out[i] = name
	}
	return out
}

// padRow returns rec resized to exactly width fields: longer rows are
// truncated, shorter rows are padded with empty strings.
func padRow(rec []string, width int) []string {
	if len(rec) == width {
		return rec
	}
	row := make([]string, width)
	copy(row, rec)
	return row
}
