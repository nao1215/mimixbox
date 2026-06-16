package printutils

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, c *Command, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := c.Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLprThenLpqThenLpd(t *testing.T) {
	spool := filepath.Join(t.TempDir(), "spool")
	doc := writeFile(t, "page content")

	// Enqueue a file.
	if _, _, err := run(t, NewLpr(), "", "-S", spool, doc); err != nil {
		t.Fatalf("lpr file: %v", err)
	}
	// Enqueue stdin.
	if _, _, err := run(t, NewLpr(), "from stdin", "-S", spool); err != nil {
		t.Fatalf("lpr stdin: %v", err)
	}

	// Queue should list two jobs.
	out, _, err := run(t, NewLpq(), "", "-S", spool)
	if err != nil {
		t.Fatalf("lpq: %v", err)
	}
	if !strings.Contains(out, "doc.txt") || !strings.Contains(out, "(stdin)") {
		t.Errorf("lpq missing jobs:\n%s", out)
	}
	if strings.Count(out, "\n") < 3 { // header + 2 rows
		t.Errorf("lpq expected 2 jobs:\n%s", out)
	}

	// Drain to an output directory.
	printed := filepath.Join(t.TempDir(), "printed")
	out, _, err = run(t, NewLpd(), "", "-S", spool, "-o", printed)
	if err != nil {
		t.Fatalf("lpd: %v", err)
	}
	if !strings.Contains(out, "printed job 1") || !strings.Contains(out, "printed job 2") {
		t.Errorf("lpd output unexpected:\n%s", out)
	}

	// Printed files should exist with the right content.
	entries, _ := os.ReadDir(printed)
	if len(entries) != 2 {
		t.Fatalf("expected 2 printed files, got %d", len(entries))
	}
	found := false
	for _, e := range entries {
		data, _ := os.ReadFile(filepath.Join(printed, e.Name()))
		if strings.Contains(string(data), "page content") {
			found = true
		}
	}
	if !found {
		t.Error("printed output missing original content")
	}

	// Queue should now be empty.
	out, _, _ = run(t, NewLpq(), "", "-S", spool)
	if !strings.Contains(out, "no entries") {
		t.Errorf("queue not drained:\n%s", out)
	}
}

func TestLpqEmpty(t *testing.T) {
	out, _, err := run(t, NewLpq(), "", "-S", filepath.Join(t.TempDir(), "empty"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "no entries" {
		t.Errorf("empty queue = %q", out)
	}
}

func TestLpdRequiresOutput(t *testing.T) {
	if _, _, err := run(t, NewLpd(), "", "-S", t.TempDir()); err == nil {
		t.Fatal("expected error without -o")
	}
}

// TestLpqCorruptControlFile proves a control file with invalid JSON is reported
// as a corrupt-spool error rather than silently skipped.
func TestLpqCorruptControlFile(t *testing.T) {
	spool := filepath.Join(t.TempDir(), "spool")
	if err := os.MkdirAll(spool, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(spool, "cf0001"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, NewLpq(), "", "-S", spool)
	if err == nil {
		t.Fatal("expected error for corrupt control file")
	}
	if !strings.Contains(errOut, "corrupt control file") {
		t.Errorf("stderr = %q, want 'corrupt control file'", errOut)
	}
}

// TestLpdMissingDataFile proves draining a job whose data file has vanished
// fails and leaves the job's control file in the queue (it is not dropped).
func TestLpdMissingDataFile(t *testing.T) {
	spool := filepath.Join(t.TempDir(), "spool")
	if _, _, err := run(t, NewLpr(), "body", "-S", spool); err != nil {
		t.Fatal(err)
	}
	jobs, err := openSpool(spool).list()
	if err != nil || len(jobs) != 1 {
		t.Fatalf("setup queue = %+v err=%v", jobs, err)
	}
	// Remove the data file out from under the spool.
	if err := os.Remove(filepath.Join(spool, jobs[0].DataFile)); err != nil {
		t.Fatal(err)
	}
	printed := filepath.Join(t.TempDir(), "printed")
	if _, _, err := run(t, NewLpd(), "", "-S", spool, "-o", printed); err == nil {
		t.Fatal("expected error draining a job with a missing data file")
	}
	// The job's control file must remain so the failure is not silently lost.
	after, err := openSpool(spool).list()
	if err != nil || len(after) != 1 {
		t.Errorf("queue after failed drain = %+v err=%v, want the job retained", after, err)
	}
}

// TestLpdPartialDrainFailure proves a drain with one good and one broken job
// prints the good one (and removes it) while reporting overall failure and
// keeping the broken job queued.
func TestLpdPartialDrainFailure(t *testing.T) {
	spool := filepath.Join(t.TempDir(), "spool")
	if _, _, err := run(t, NewLpr(), "good one", "-S", spool); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, NewLpr(), "broken one", "-S", spool); err != nil {
		t.Fatal(err)
	}
	jobs, err := openSpool(spool).list()
	if err != nil || len(jobs) != 2 {
		t.Fatalf("setup queue = %+v err=%v", jobs, err)
	}
	// Break the second job by deleting its data file.
	if err := os.Remove(filepath.Join(spool, jobs[1].DataFile)); err != nil {
		t.Fatal(err)
	}
	printed := filepath.Join(t.TempDir(), "printed")
	out, _, err := run(t, NewLpd(), "", "-S", spool, "-o", printed)
	if err == nil {
		t.Fatal("expected overall failure when one job cannot be drained")
	}
	if !strings.Contains(out, "printed job 1") {
		t.Errorf("good job should still print: %q", out)
	}
	// Exactly the good job's output exists.
	entries, _ := os.ReadDir(printed)
	if len(entries) != 1 {
		t.Errorf("expected 1 printed file, got %d", len(entries))
	}
	// The broken job remains queued; the good job was removed.
	after, err := openSpool(spool).list()
	if err != nil || len(after) != 1 || after[0].ID != jobs[1].ID {
		t.Errorf("queue after partial drain = %+v err=%v, want only the broken job", after, err)
	}
}

func TestIDsIncrement(t *testing.T) {
	spool := filepath.Join(t.TempDir(), "spool")
	if _, _, err := run(t, NewLpr(), "a", "-S", spool); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, NewLpr(), "b", "-S", spool); err != nil {
		t.Fatal(err)
	}
	jobs, err := openSpool(spool).list()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 || jobs[0].ID != 1 || jobs[1].ID != 2 {
		t.Errorf("ids not sequential: %+v", jobs)
	}
}
