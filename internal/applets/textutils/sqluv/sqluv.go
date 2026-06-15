//
// mimixbox/internal/applets/textutils/sqluv/sqluv.go
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

// Package sqluv implements the sqluv applet: a script-friendly SQL viewer and
// query runner over local CSV / TSV / LTSV files and SQLite3 databases. It is a
// migration of the archived nao1215/sqluv project (a terminal UI for SQL over
// RDBMS and delimited files) into a MimixBox applet.
//
// The applet has two modes:
//
//   - Headless (non-interactive): `sqluv --execute 'SELECT ...' SOURCE ...`
//     loads every SOURCE into an in-memory SQLite database (one table per
//     delimited file, plus the schema of any SQLite source), runs the SQL, and
//     writes the result as a table, CSV, TSV, or JSON. This path is fully
//     testable in CI and is what shells and LLMs should use.
//   - TUI (interactive): `sqluv SOURCE ...` with no --execute opens a minimal
//     full-screen viewer. The initial port only renders the loaded tables and
//     exits cleanly; richer keybindings are a long-term goal.
//
// All DB-backed access is read-only by default (--read-only, on by default) so
// the applet is safe to run inside MimixBox. Query history is appended to a
// configurable file (--history-file) so tests never touch the real home
// directory. HTTPS and S3 sources from the original sqluv are not migrated yet
// and fail with a deterministic, documented error.
package sqluv

import (
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the sqluv applet.
type Command struct{}

// New returns a sqluv command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sqluv" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "SQL viewer & query runner for CSV/TSV/LTSV and SQLite"
}

// options holds the parsed command-line state shared by both modes.
type options struct {
	execute     string
	output      string
	readOnly    bool
	historyFile string
}

// Run executes sqluv. When --execute is given it runs the headless query path;
// otherwise it starts the minimal TUI smoke path.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE_OR_DSN]...", stdio.Err).WithHelp(command.Help{
		Description: "Open local CSV/TSV/LTSV files and SQLite3 databases as SQL tables.\n" +
			"\n" +
			"With --execute, sqluv runs in headless mode: it loads every source into an\n" +
			"in-memory SQLite database (one table per delimited file, named after the file\n" +
			"stem; tables of a SQLite source keep their own names), runs the SQL once, and\n" +
			"prints the result. Without --execute it starts a minimal full-screen viewer\n" +
			"that lists the loaded tables and exits on 'q' or Ctrl-C.\n" +
			"\n" +
			"Supported source formats: .csv, .tsv, .ltsv, and SQLite3 database files\n" +
			"(.db/.sqlite/.sqlite3). Each may be transparently compressed with .gz, .bz2,\n" +
			".xz, or .zst (for example data.csv.gz). All database access is read-only by\n" +
			"default. HTTPS and S3 sources are NOT migrated yet and fail with a clear error.",
		Examples: []command.Example{
			{Command: "sqluv data.csv --execute 'select * from data limit 5'", Explain: "Query a CSV file as table 'data'."},
			{Command: "sqluv sample.db --execute 'select name from sqlite_master' --output=json", Explain: "Inspect a SQLite schema as JSON."},
			{Command: "sqluv access.tsv.gz --execute 'select count(*) from access'", Explain: "Query a gzip-compressed TSV file."},
			{Command: "sqluv --history-file /tmp/sqluv-history.db sample.db", Explain: "Open the TUI with an isolated history file."},
		},
		ExitStatus: "0  success.\n1  usage error, load failure, query failure, or unsupported source.",
		Notes: []string{
			"Headless mode (--execute) is implemented and CI-tested; TUI mode is a minimal smoke path only.",
			"Migrated backends: local CSV/TSV/LTSV files and SQLite3 databases.",
			"Not yet migrated (deterministic error): MySQL/PostgreSQL/SQL Server DSNs, HTTPS URLs, and S3 URLs.",
			"--read-only is on by default; pass --read-only=false to allow writes against a real SQLite file (DML still runs on the in-memory copy in --execute mode).",
		},
	})

	execute := fs.StringP("execute", "e", "", "run SQL non-interactively and print the result, then exit")
	output := fs.StringP("output", "o", "table", "headless output format: table, csv, tsv, or json")
	readOnly := fs.Bool("read-only", true, "open database sources read-only (default true)")
	historyFile := fs.String("history-file", "", "path to the query-history file (default: a temp file; never the real home dir)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		execute:     *execute,
		output:      strings.ToLower(*output),
		readOnly:    *readOnly,
		historyFile: *historyFile,
	}

	sources := fs.Args()

	// Validate every source up front so unsupported source types fail with a
	// deterministic error before any work begins.
	for _, src := range sources {
		if err := validateSource(src); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
			return command.SilentFailure()
		}
	}

	if opts.execute != "" {
		if err := runHeadless(stdio, sources, opts); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
			return command.SilentFailure()
		}
		return nil
	}

	if len(sources) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: missing operand: specify at least one FILE_OR_DSN, or use --execute\n", c.Name())
		_, _ = fmt.Fprintf(stdio.Err, "Try '%s --help' for more information.\n", c.Name())
		return command.SilentFailure()
	}

	if err := runTUI(stdio, sources, opts); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}
	return nil
}
