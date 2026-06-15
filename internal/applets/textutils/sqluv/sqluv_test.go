package sqluv_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/nao1215/mimixbox/internal/applets/textutils/sqluv"
	"github.com/nao1215/mimixbox/internal/command"
	"github.com/ulikunitz/xz"
	_ "modernc.org/sqlite"
)

// run executes sqluv with the given stdin and arguments and returns stdout,
// stderr, and the error.
func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := sqluv.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, path, content string) string {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

const csvFixture = "id,name\n1,alice\n2,bob\n3,carol\n"

// ----- format detection / source classification -----

func TestClassifySource(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want sqluv.SourceKindAlias
	}{
		{"data.csv", sqluv.KindDelimited},
		{"data.tsv", sqluv.KindDelimited},
		{"data.ltsv", sqluv.KindDelimited},
		{"data.csv.gz", sqluv.KindDelimited},
		{"path/to/access.tsv.zst", sqluv.KindDelimited},
		{"sample.db", sqluv.KindSQLite},
		{"sample.sqlite3", sqluv.KindSQLite},
		{"mydata", sqluv.KindSQLite},
		{"https://example.com/data.csv", sqluv.KindUnsupported},
		{"s3://bucket/data.csv", sqluv.KindUnsupported},
		{"user@tcp(127.0.0.1:3306)/db", sqluv.KindUnsupported},
		{"postgres://u:p@host/db", sqluv.KindUnsupported},
	}
	for _, tc := range cases {
		if got := sqluv.ClassifySource(tc.in); got != tc.want {
			t.Errorf("classifySource(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestDetectFormatByName(t *testing.T) {
	t.Parallel()
	cases := map[string]sqluv.FileFormatAlias{
		"a.csv":  sqluv.FormatCSV,
		"a.tsv":  sqluv.FormatTSV,
		"a.ltsv": sqluv.FormatLTSV,
		"a.txt":  sqluv.FormatUnknown,
		"a.db":   sqluv.FormatUnknown,
	}
	for in, want := range cases {
		if got := sqluv.DetectFormatByName(in); got != want {
			t.Errorf("detectFormatByName(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestSplitCompression(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in       string
		wantBase string
		wantComp sqluv.CompressionAlias
	}{
		{"data.csv", "data.csv", sqluv.CompNone},
		{"data.csv.gz", "data.csv", sqluv.CompGzip},
		{"data.csv.bz2", "data.csv", sqluv.CompBzip2},
		{"data.csv.xz", "data.csv", sqluv.CompXz},
		{"data.csv.zst", "data.csv", sqluv.CompZstd},
	}
	for _, tc := range cases {
		base, comp := sqluv.SplitCompression(tc.in)
		if base != tc.wantBase || comp != tc.wantComp {
			t.Errorf("splitCompression(%q) = (%q,%v), want (%q,%v)", tc.in, base, comp, tc.wantBase, tc.wantComp)
		}
	}
}

func TestTableNameFor(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"data.csv":             "data",
		"path/to/access.tsv":   "access",
		"weird name.csv":       "weird_name",
		"123start.csv":         "t_123start",
		"logs.ltsv.gz":         "logs",
	}
	for in, want := range cases {
		if got := sqluv.TableNameFor(in); got != want {
			t.Errorf("tableNameFor(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsMutating(t *testing.T) {
	t.Parallel()
	mutating := []string{"INSERT INTO t VALUES(1)", "delete from t", "  UPDATE t SET a=1", "DROP TABLE t", "create table x(a)"}
	for _, q := range mutating {
		if !sqluv.IsMutating(q) {
			t.Errorf("isMutating(%q) = false, want true", q)
		}
	}
	readonly := []string{"SELECT * FROM t", "  select 1", "WITH cte AS (SELECT 1) SELECT * FROM cte"}
	for _, q := range readonly {
		if sqluv.IsMutating(q) {
			t.Errorf("isMutating(%q) = true, want false", q)
		}
	}
}

func TestValidateOutputFormat(t *testing.T) {
	t.Parallel()
	for _, ok := range []string{"table", "csv", "tsv", "json"} {
		if err := sqluv.ValidateOutputFmt(ok); err != nil {
			t.Errorf("validateOutputFormat(%q) error = %v", ok, err)
		}
	}
	if err := sqluv.ValidateOutputFmt("yaml"); err == nil {
		t.Error("validateOutputFormat(yaml) = nil, want error")
	}
}

func TestHistoryPathDefaultsToTemp(t *testing.T) {
	t.Parallel()
	def := sqluv.HistoryPath("")
	if !strings.HasPrefix(def, os.TempDir()) {
		t.Errorf("default history path %q not under temp dir %q", def, os.TempDir())
	}
	if got := sqluv.HistoryPath("/tmp/custom-history.log"); got != "/tmp/custom-history.log" {
		t.Errorf("explicit history path = %q", got)
	}
}

// ----- loader unit tests (delimited + compression) -----

func TestLoadDelimitedCSV(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "data.csv"), csvFixture)
	tbl, err := sqluv.LoadDelimitedForTst(path)
	if err != nil {
		t.Fatalf("load csv: %v", err)
	}
	if tbl.TableName() != "data" {
		t.Errorf("table name = %q, want data", tbl.TableName())
	}
	if strings.Join(tbl.Columns(), ",") != "id,name" {
		t.Errorf("columns = %v", tbl.Columns())
	}
	if len(tbl.Rows()) != 3 {
		t.Errorf("rows = %d, want 3", len(tbl.Rows()))
	}
}

func TestLoadDelimitedTSV(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "data.tsv"), "id\tname\n1\talice\n2\tbob\n")
	tbl, err := sqluv.LoadDelimitedForTst(path)
	if err != nil {
		t.Fatalf("load tsv: %v", err)
	}
	if len(tbl.Rows()) != 2 || tbl.Rows()[0][1] != "alice" {
		t.Errorf("tsv rows = %v", tbl.Rows())
	}
}

func TestLoadDelimitedLTSV(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	content := "id:1\tname:alice\nid:2\tname:bob\textra:x\n"
	path := writeFile(t, filepath.Join(dir, "logs.ltsv"), content)
	tbl, err := sqluv.LoadDelimitedForTst(path)
	if err != nil {
		t.Fatalf("load ltsv: %v", err)
	}
	if strings.Join(tbl.Columns(), ",") != "id,name,extra" {
		t.Errorf("ltsv columns = %v", tbl.Columns())
	}
	if len(tbl.Rows()) != 2 {
		t.Errorf("ltsv rows = %d, want 2", len(tbl.Rows()))
	}
	// first row has no "extra" -> empty string
	if tbl.Rows()[0][2] != "" {
		t.Errorf("ltsv row0 extra = %q, want empty", tbl.Rows()[0][2])
	}
}

func TestLoadDelimitedCompressed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	gzPath := filepath.Join(dir, "data.csv.gz")
	writeGzip(t, gzPath, csvFixture)
	xzPath := filepath.Join(dir, "data.csv.xz")
	writeXz(t, xzPath, csvFixture)
	zstPath := filepath.Join(dir, "data.csv.zst")
	writeZstd(t, zstPath, csvFixture)

	for _, p := range []string{gzPath, xzPath, zstPath} {
		tbl, err := sqluv.LoadDelimitedForTst(p)
		if err != nil {
			t.Fatalf("load %s: %v", p, err)
		}
		if len(tbl.Rows()) != 3 {
			t.Errorf("%s rows = %d, want 3", p, len(tbl.Rows()))
		}
	}
}

// ----- headless execution end-to-end -----

func TestHeadlessCSVQuery(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "data.csv"), csvFixture)
	hist := filepath.Join(dir, "history.log")

	out, errStr, err := run(t, "",
		"--history-file", hist, "--output", "csv",
		"--execute", "select name from data order by id limit 2", path)
	if err != nil {
		t.Fatalf("run error = %v, stderr=%s", err, errStr)
	}
	want := "name\nalice\nbob\n"
	if out != want {
		t.Errorf("csv output = %q, want %q", out, want)
	}
	// history was recorded to the temp file, not the home dir.
	if _, err := os.Stat(hist); err != nil {
		t.Errorf("history file not written: %v", err)
	}
}

func TestHeadlessJSONOutput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "data.csv"), csvFixture)
	out, errStr, err := run(t, "",
		"--history-file", filepath.Join(dir, "h.log"), "--output", "json",
		"--execute", "select id,name from data where id=1", path)
	if err != nil {
		t.Fatalf("run error = %v, stderr=%s", err, errStr)
	}
	var got []map[string]string
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("json unmarshal: %v (out=%q)", err, out)
	}
	if len(got) != 1 || got[0]["name"] != "alice" {
		t.Errorf("json = %v", got)
	}
}

func TestHeadlessTableOutput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "data.csv"), csvFixture)
	out, _, err := run(t, "",
		"--history-file", filepath.Join(dir, "h.log"),
		"--execute", "select count(*) as n from data", path)
	if err != nil {
		t.Fatalf("run error = %v", err)
	}
	if !strings.Contains(out, "| n") || !strings.Contains(out, "(1 row)") {
		t.Errorf("table output = %q", out)
	}
}

func TestHeadlessSQLiteSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "sample.db")
	makeSQLiteFixture(t, dbPath)

	out, errStr, err := run(t, "",
		"--history-file", filepath.Join(dir, "h.log"), "--output", "csv",
		"--execute", "select title from books order by id", dbPath)
	if err != nil {
		t.Fatalf("run error = %v, stderr=%s", err, errStr)
	}
	if !strings.Contains(out, "go-in-action") || !strings.Contains(out, "title") {
		t.Errorf("sqlite query output = %q", out)
	}
}

func TestHeadlessReadOnlyRejectsMutation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "data.csv"), csvFixture)
	_, errStr, err := run(t, "",
		"--history-file", filepath.Join(dir, "h.log"),
		"--execute", "delete from data", path)
	if err == nil {
		t.Fatal("expected read-only error for delete")
	}
	if !strings.Contains(errStr, "read-only") {
		t.Errorf("stderr = %q, want read-only message", errStr)
	}
}

func TestHeadlessUnsupportedSource(t *testing.T) {
	t.Parallel()
	_, errStr, err := run(t, "",
		"--execute", "select 1", "https://example.com/data.csv")
	if err == nil {
		t.Fatal("expected error for https source")
	}
	if !strings.Contains(errStr, "HTTPS sources are not migrated") {
		t.Errorf("stderr = %q, want HTTPS unsupported message", errStr)
	}
}

func TestHeadlessBadOutputFormat(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "data.csv"), csvFixture)
	_, errStr, err := run(t, "",
		"--output", "xml", "--execute", "select 1", path)
	if err == nil {
		t.Fatal("expected error for bad output format")
	}
	if !strings.Contains(errStr, "unsupported --output format") {
		t.Errorf("stderr = %q", errStr)
	}
}

// ----- TUI smoke (non-terminal stdin must render and exit) -----

func TestTUISmokeExitsCleanly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeFile(t, filepath.Join(dir, "data.csv"), csvFixture)
	// stdin "q\n" makes the non-terminal loop terminate immediately.
	out, errStr, err := run(t, "q\n",
		"--history-file", filepath.Join(dir, "h.log"), path)
	if err != nil {
		t.Fatalf("tui smoke error = %v, stderr=%s", err, errStr)
	}
	if !strings.Contains(out, "minimal viewer") || !strings.Contains(out, "data") || !strings.Contains(out, "bye") {
		t.Errorf("tui output = %q", out)
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, errStr, err := run(t, "")
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if !strings.Contains(errStr, "missing operand") {
		t.Errorf("stderr = %q", errStr)
	}
}

func TestHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: sqluv") || !strings.Contains(out, "headless") {
		t.Errorf("--help out = %q", out)
	}
	out, _, err = run(t, "", "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "sqluv (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}

// ----- fixture helpers -----

func writeGzip(t *testing.T, path, content string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	w := gzip.NewWriter(f)
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeXz(t *testing.T, path, content string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	w, err := xz.NewWriter(f)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeZstd(t *testing.T, path, content string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	w, err := zstd.NewWriter(f)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func makeSQLiteFixture(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	stmts := []string{
		"CREATE TABLE books (id INTEGER PRIMARY KEY, title TEXT)",
		"INSERT INTO books (id, title) VALUES (1, 'go-in-action'), (2, 'the-go-programming-language')",
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("exec %q: %v", s, err)
		}
	}
}
