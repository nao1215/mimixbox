package expand_test

import (
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/expand"
)

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := expand.New()
	if c.Name() != "expand" {
		t.Errorf("Name() = %q, want %q", c.Name(), "expand")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestRunTwoMissingFiles drives the keep() "an error is already recorded"
// branch by failing on two files in a row; the first failure must be the one
// returned even though a second failure also occurred.
func TestRunTwoMissingFiles(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/a", "/no/such/b")
	if err == nil {
		t.Fatal("expected error for two missing files")
	}
	// The first failing path must be reported before the second one, proving the
	// returned error corresponds to the first failure (a.before b). The returned
	// error itself is an intentionally silent failure whose message is already on
	// stderr, so assert ordering on stderr rather than on err.Error().
	idxA := strings.Index(errOut, "/no/such/a")
	idxB := strings.Index(errOut, "/no/such/b")
	if idxA < 0 || idxB < 0 {
		t.Errorf("stderr = %q, want both missing files reported", errOut)
	} else if idxA > idxB {
		t.Errorf("stderr = %q, want /no/such/a reported as the first failure", errOut)
	}
}

// TestRunNonPositiveTabFallsBackToEight checks that -t 0 (and negatives) fall
// back to the default 8-column tab stop instead of dividing by zero.
func TestRunNonPositiveTabFallsBackToEight(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\tb\n", "-t", "0")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "a       b\n" { // single 'a' then advance to column 8.
		t.Errorf("out = %q, want default 8-column expansion", out)
	}
}

// TestRunInitialPreservesTabsAfterText exercises the -i path where a tab follows
// a non-blank: leading tabs expand but the embedded tab is preserved verbatim.
func TestRunInitialPreservesTabsAfterText(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "\tx\ty\n", "-i", "-t", "4")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "    x\ty\n" {
		t.Errorf("out = %q, want leading tab expanded and embedded tab kept", out)
	}
}
