// mimixbox/cmd/mimixbox/list_suggest_test.go
//
// Tests for the top-level UX surfaces added in issues #781-#783: `--list --json`,
// `--list` filtering, and nearest-name suggestions for unknown commands.
package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestRunListJSON(t *testing.T) {
	io, out, _ := newIO()
	if code := run([]string{"mimixbox", "--list", "--json"}, io); code != command.ExitSuccess {
		t.Fatalf("exit = %d, want 0", code)
	}
	var got []map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("--list --json did not emit valid JSON: %v\n%s", err, out.String())
	}
	if len(got) == 0 {
		t.Fatal("--list --json emitted an empty array")
	}
	// Schema: every object has the four documented keys.
	for _, e := range got {
		for _, k := range []string{"name", "synopsis", "subsystem", "stability"} {
			if _, ok := e[k]; !ok {
				t.Errorf("JSON object missing key %q: %v", k, e)
			}
		}
	}
	// Must include cat and ls, and be sorted by name.
	var names []string
	for _, e := range got {
		names = append(names, e["name"].(string))
	}
	joined := "," + strings.Join(names, ",") + ","
	for _, must := range []string{"cat", "ls"} {
		if !strings.Contains(joined, ","+must+",") {
			t.Errorf("--list --json is missing %q", must)
		}
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("--list --json not sorted: %q before %q", names[i-1], names[i])
		}
	}
}

func TestRunListFilterPrefix(t *testing.T) {
	io, out, _ := newIO()
	if code := run([]string{"mimixbox", "--list", "--filter=cat"}, io); code != command.ExitSuccess {
		t.Fatalf("exit = %d, want 0", code)
	}
	got := out.String()
	if !strings.Contains(got, "cat") {
		t.Errorf("--filter=cat should include cat, got %q", got)
	}
	// "ls" is not cat-prefixed and must be excluded. Guard against substring
	// false-positives by checking word boundaries via the listing layout.
	for _, line := range strings.Split(strings.TrimSpace(got), "\n") {
		name := strings.TrimSpace(strings.SplitN(line, " - ", 2)[0])
		if name != "" && !strings.HasPrefix(name, "cat") {
			t.Errorf("--filter=cat returned non-matching applet %q", name)
		}
	}
}

func TestRunListBareGlobOperand(t *testing.T) {
	io, out, _ := newIO()
	if code := run([]string{"mimixbox", "--list", "cat*"}, io); code != command.ExitSuccess {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "cat") {
		t.Errorf("--list cat* should include cat, got %q", out.String())
	}
}

func TestRunListSubsystemFilter(t *testing.T) {
	io, out, _ := newIO()
	if code := run([]string{"mimixbox", "--list", "--subsystem=textutils"}, io); code != command.ExitSuccess {
		t.Fatalf("exit = %d, want 0", code)
	}
	got := out.String()
	if !strings.Contains(got, "cat") {
		t.Errorf("--subsystem=textutils should include cat, got %q", got)
	}
	// ls is in fileutils, so it must not appear as a listed applet name.
	for _, line := range strings.Split(strings.TrimSpace(got), "\n") {
		name := strings.TrimSpace(strings.SplitN(line, " - ", 2)[0])
		if name == "ls" {
			t.Errorf("--subsystem=textutils must not include ls")
		}
	}
}

func TestRunListUnknownOption(t *testing.T) {
	io, out, errBuf := newIO()
	if code := run([]string{"mimixbox", "--list", "--bogus"}, io); code != command.ExitFailure {
		t.Fatalf("exit = %d, want 1", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout must stay clean, got %q", out.String())
	}
	if !strings.Contains(errBuf.String(), "unknown option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRunUnknownCommandSuggests(t *testing.T) {
	io, out, errBuf := newIO()
	if code := run([]string{"mimixbox", "lss"}, io); code != command.ExitFailure {
		t.Fatalf("exit = %d, want 1", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout must stay clean on error, got %q", out.String())
	}
	stderr := errBuf.String()
	// Error-first concise suggestion line before the full wall.
	if !strings.Contains(stderr, "'lss' is not a mimixbox command.") {
		t.Errorf("stderr missing error-first line: %q", stderr)
	}
	if !strings.Contains(stderr, "Did you mean:") || !strings.Contains(stderr, "ls") {
		t.Errorf("stderr missing 'Did you mean: ls': %q", stderr)
	}
	// The suggestion must come before the full applet wall.
	idxSuggest := strings.Index(stderr, "Did you mean:")
	idxWall := strings.Index(stderr, "[Commands supported by MimixBox]")
	if idxSuggest < 0 || idxWall < 0 || idxSuggest > idxWall {
		t.Errorf("suggestion should precede the full wall: %q", stderr)
	}
}

func TestRunUnknownOptionStillReported(t *testing.T) {
	// An unknown --option should NOT get applet suggestions (it is not a
	// command name), but must still be reported on stderr.
	io, _, errBuf := newIO()
	if code := run([]string{"mimixbox", "--definitely-not-an-option"}, io); code != command.ExitFailure {
		t.Fatalf("exit = %d, want 1", code)
	}
	stderr := errBuf.String()
	if !strings.Contains(stderr, "is not a mimixbox command or option") {
		t.Errorf("stderr = %q", stderr)
	}
	if strings.Contains(stderr, "Did you mean:") {
		t.Errorf("options should not get command suggestions: %q", stderr)
	}
}
