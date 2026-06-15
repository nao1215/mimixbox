//
// mimixbox/internal/applets/textutils/sqluv/source.go
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
	"fmt"
	"path/filepath"
	"strings"
)

// sourceKind classifies a command-line source operand.
type sourceKind int

const (
	// kindDelimited is a local CSV/TSV/LTSV file (possibly compressed).
	kindDelimited sourceKind = iota
	// kindSQLite is a local SQLite3 database file.
	kindSQLite
	// kindUnsupported is a source type recognized by the original sqluv but not
	// yet migrated (HTTPS, S3, or a remote RDBMS DSN).
	kindUnsupported
)

// fileFormat is the delimited-file format of a source.
type fileFormat int

const (
	formatUnknown fileFormat = iota
	formatCSV
	formatTSV
	formatLTSV
)

func (f fileFormat) String() string {
	switch f {
	case formatCSV:
		return "csv"
	case formatTSV:
		return "tsv"
	case formatLTSV:
		return "ltsv"
	default:
		return "unknown"
	}
}

// compression is the transparent compression wrapping a delimited source.
type compression int

const (
	compNone compression = iota
	compGzip
	compBzip2
	compXz
	compZstd
)

// classifySource decides how a source operand should be handled. It inspects
// the operand string only (scheme and extension); it does not touch the
// filesystem, which keeps classification pure and testable.
func classifySource(src string) sourceKind {
	lower := strings.ToLower(src)
	switch {
	case strings.HasPrefix(lower, "https://"),
		strings.HasPrefix(lower, "http://"),
		strings.HasPrefix(lower, "s3://"),
		isRemoteDSN(lower):
		return kindUnsupported
	}

	base, _ := splitCompression(src)
	switch detectFormatByName(base) {
	case formatCSV, formatTSV, formatLTSV:
		return kindDelimited
	}
	if isSQLiteName(base) {
		return kindSQLite
	}
	// Anything else with no recognizable delimited extension is treated as a
	// SQLite database file (the original sqluv accepts arbitrary *.db names).
	return kindSQLite
}

// isRemoteDSN reports whether src looks like a remote-RDBMS DSN/URL that the
// original sqluv supported (MySQL, PostgreSQL, SQL Server) but that this port
// has not migrated yet.
func isRemoteDSN(lower string) bool {
	prefixes := []string{
		"mysql://", "mysql:", "postgres://", "postgresql://",
		"sqlserver://", "mssql://", "tcp(", "user:", "host=",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	// "user@tcp(...)" style MySQL DSNs.
	if strings.Contains(lower, "@tcp(") || strings.Contains(lower, "@unix(") {
		return true
	}
	return false
}

// validateSource returns a deterministic, documented error for source types
// that are recognized but not migrated, and nil for supported sources.
func validateSource(src string) error {
	if classifySource(src) != kindUnsupported {
		return nil
	}
	lower := strings.ToLower(src)
	switch {
	case strings.HasPrefix(lower, "https://"), strings.HasPrefix(lower, "http://"):
		return fmt.Errorf("HTTPS sources are not migrated yet: %q (only local files and SQLite are supported in this port)", src)
	case strings.HasPrefix(lower, "s3://"):
		return fmt.Errorf("S3 sources are not migrated yet: %q (only local files and SQLite are supported in this port)", src)
	default:
		return fmt.Errorf("remote RDBMS DSNs (MySQL/PostgreSQL/SQL Server) are not migrated yet: %q (only local files and SQLite are supported in this port)", src)
	}
}

// splitCompression strips a trailing compression extension from name and
// reports which compression it implies. The returned base has the compression
// suffix removed so the underlying format can still be detected.
func splitCompression(name string) (base string, comp compression) {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".gz"):
		return name[:len(name)-3], compGzip
	case strings.HasSuffix(lower, ".bz2"):
		return name[:len(name)-4], compBzip2
	case strings.HasSuffix(lower, ".xz"):
		return name[:len(name)-3], compXz
	case strings.HasSuffix(lower, ".zst"):
		return name[:len(name)-4], compZstd
	default:
		return name, compNone
	}
}

// detectFormatByName guesses the delimited format from a file name's extension.
func detectFormatByName(name string) fileFormat {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".csv":
		return formatCSV
	case ".tsv":
		return formatTSV
	case ".ltsv":
		return formatLTSV
	default:
		return formatUnknown
	}
}

// isSQLiteName reports whether name has a conventional SQLite database
// extension.
func isSQLiteName(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".db", ".sqlite", ".sqlite3", ".sqlite2":
		return true
	default:
		return false
	}
}

// tableNameFor derives a SQL table name from a delimited source path: the file
// stem with the compression and format extensions removed and non-identifier
// characters replaced by underscores.
func tableNameFor(src string) string {
	base := filepath.Base(src)
	base, _ = splitCompression(base)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return sanitizeIdentifier(base)
}

// sanitizeIdentifier turns an arbitrary string into a safe SQL identifier:
// leading digits are prefixed with "t_", and any character that is not a
// letter, digit, or underscore becomes an underscore.
func sanitizeIdentifier(s string) string {
	if s == "" {
		return "t"
	}
	var b strings.Builder
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			if i == 0 {
				b.WriteString("t_")
			}
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" {
		return "t"
	}
	return out
}
