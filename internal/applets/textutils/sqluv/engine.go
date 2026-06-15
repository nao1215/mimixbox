//
// mimixbox/internal/applets/textutils/sqluv/engine.go
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
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (no cgo)
)

// resultSet is the outcome of a query: ordered columns and string-rendered
// rows (NULLs become the empty string, matching the original sqluv export).
type resultSet struct {
	columns []string
	rows    [][]string
}

// engine wraps an in-memory SQLite database that holds every loaded source. A
// delimited file becomes one table; a SQLite source's tables are copied in by
// ATTACHing the file read-only and creating in-memory copies so the original
// file is never modified.
type engine struct {
	db *sql.DB
}

// newEngine opens a fresh in-memory SQLite database.
func newEngine() (*engine, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("open in-memory database: %w", err)
	}
	return &engine{db: db}, nil
}

// close releases the underlying database.
func (e *engine) close() error { return e.db.Close() }

// loadTable creates a table from a delimited source. Every column is TEXT,
// which matches how a delimited file is untyped on disk.
func (e *engine) loadTable(t *table) error {
	cols := make([]string, len(t.columns))
	for i, c := range t.columns {
		cols[i] = quoteIdent(c) + " TEXT"
	}
	create := fmt.Sprintf("CREATE TABLE %s (%s)", quoteIdent(t.name), strings.Join(cols, ", "))
	if _, err := e.db.Exec(create); err != nil {
		return fmt.Errorf("create table %q: %w", t.name, err)
	}

	if len(t.rows) == 0 {
		return nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(t.columns)), ",")
	insert := fmt.Sprintf("INSERT INTO %s VALUES (%s)", quoteIdent(t.name), placeholders)

	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(insert)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, row := range t.rows {
		args := make([]any, len(row))
		for i, v := range row {
			args[i] = v
		}
		if _, err := stmt.Exec(args...); err != nil {
			_ = stmt.Close()
			_ = tx.Rollback()
			return fmt.Errorf("insert into %q: %w", t.name, err)
		}
	}
	_ = stmt.Close()
	return tx.Commit()
}

// loadSQLiteFile copies every user table of a SQLite database file into the
// in-memory database. The file is attached in read-only mode so it is never
// modified, regardless of the --read-only flag.
func (e *engine) loadSQLiteFile(path string) error {
	// ATTACH the file read-only via a URI filename with mode=ro.
	attach := fmt.Sprintf("ATTACH DATABASE 'file:%s?mode=ro&immutable=1' AS src", escapeSingleQuotes(path))
	if _, err := e.db.Exec(attach); err != nil {
		return fmt.Errorf("attach sqlite file %q: %w", path, err)
	}
	defer func() { _, _ = e.db.Exec("DETACH DATABASE src") }()

	rows, err := e.db.Query(`SELECT name FROM src.sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return fmt.Errorf("read schema of %q: %w", path, err)
	}
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = rows.Close()
			return err
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()

	for _, name := range names {
		copySQL := fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM src.%s", quoteIdent(name), quoteIdent(name))
		if _, err := e.db.Exec(copySQL); err != nil {
			return fmt.Errorf("copy table %q from %q: %w", name, path, err)
		}
	}
	return nil
}

// query runs sql against the in-memory database and returns the result as
// strings. When readOnly is set, statements that would modify data are rejected
// before execution.
func (e *engine) query(sqlText string, readOnly bool) (*resultSet, error) {
	if readOnly && isMutating(sqlText) {
		return nil, fmt.Errorf("refusing to run a data-modifying statement in read-only mode (pass --read-only=false to allow): %s", firstWord(sqlText))
	}

	rows, err := e.db.Query(sqlText)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	rs := &resultSet{columns: cols}
	for rows.Next() {
		raw := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make([]string, len(cols))
		for i, v := range raw {
			row[i] = renderValue(v)
		}
		rs.rows = append(rs.rows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rs, nil
}

// renderValue converts a scanned SQL value into its string form. NULL becomes
// the empty string so output formats stay rectangular.
func renderValue(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(x)
	case string:
		return x
	default:
		return fmt.Sprintf("%v", x)
	}
}

// isMutating reports whether sql begins with a statement keyword that would
// modify data or schema. It is a conservative guard for read-only mode.
func isMutating(sqlText string) bool {
	switch strings.ToUpper(firstWord(sqlText)) {
	case "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
		"REPLACE", "TRUNCATE", "PRAGMA", "ATTACH", "DETACH", "VACUUM", "REINDEX":
		return true
	default:
		return false
	}
}

// firstWord returns the first whitespace-delimited token of s, ignoring leading
// whitespace and a leading "WITH" CTE keyword's noise is intentionally treated
// as non-mutating since the final statement governs.
func firstWord(s string) string {
	fields := strings.Fields(strings.TrimSpace(s))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// quoteIdent wraps a SQL identifier in double quotes, escaping embedded quotes.
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// escapeSingleQuotes doubles single quotes for safe embedding in a string
// literal used by ATTACH.
func escapeSingleQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
