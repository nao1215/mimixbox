package unshadow

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// errWriter fails every Write, to exercise merge's write-error branch.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

// TestParseShadowSkipsBlankAndShortLines covers the empty-line and short-field
// branches of parseShadow.
func TestParseShadowSkipsBlankAndShortLines(t *testing.T) {
	t.Parallel()
	in := strings.NewReader(
		"root:$6$HASH:19000:0:99999:7:::\n" + // normal
			"\n" + // blank line, skipped
			"noco lon line\n" + // no colon -> SplitN gives 1 field, skipped
			"u:hash\n", // exactly two fields
	)
	hashes, err := parseShadow(in)
	if err != nil {
		t.Fatalf("parseShadow error = %v", err)
	}
	if hashes["root"] != "$6$HASH" {
		t.Errorf("root hash = %q, want $6$HASH", hashes["root"])
	}
	if hashes["u"] != "hash" {
		t.Errorf("u hash = %q, want hash", hashes["u"])
	}
	if _, ok := hashes["noco lon line"]; ok {
		t.Error("a line without a colon must not produce an entry")
	}
	if len(hashes) != 2 {
		t.Errorf("hashes has %d entries, want 2", len(hashes))
	}
}

// TestMergeSkipsBlankAndShortLines covers the empty-line and short-field guards
// in merge.
func TestMergeSkipsBlankAndShortLines(t *testing.T) {
	t.Parallel()
	passwd := strings.NewReader(
		"root:x:0:0:root:/root:/bin/bash\n" +
			"\n" + // blank line, skipped
			"justname\n" + // single field, skipped
			"alice:x:1000:1000::/home/alice:/bin/sh\n",
	)
	hashes := map[string]string{"root": "$6$ROOT", "alice": "$6$ALICE"}

	var out bytes.Buffer
	c := &Command{}
	if err := c.merge(command.IO{Out: &out}, passwd, hashes); err != nil {
		t.Fatalf("merge error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "root:$6$ROOT:0:0:root:/root:/bin/bash") {
		t.Errorf("root not merged: %q", got)
	}
	if !strings.Contains(got, "alice:$6$ALICE:1000:1000") {
		t.Errorf("alice not merged: %q", got)
	}
	if strings.Contains(got, "justname") {
		t.Errorf("single-field line should have been skipped: %q", got)
	}
}

// TestMergeWriteError covers the Fprintln error branch of merge.
func TestMergeWriteError(t *testing.T) {
	t.Parallel()
	passwd := strings.NewReader("root:x:0:0::/root:/bin/sh\n")
	c := &Command{}
	err := c.merge(command.IO{Out: errWriter{}}, passwd, map[string]string{})
	if err == nil {
		t.Fatal("expected a write error to propagate")
	}
}
