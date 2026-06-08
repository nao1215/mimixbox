package version

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrint(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	Print(&b, "cat")
	got := b.String()
	if !strings.HasPrefix(got, "cat (mimixbox) ") {
		t.Errorf("Print output = %q", got)
	}
	if !strings.Contains(got, Version) {
		t.Errorf("Print output %q does not contain version %q", got, Version)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("Print output should end with a newline: %q", got)
	}
}

func TestVersionNotEmpty(t *testing.T) {
	t.Parallel()
	if Version == "" {
		t.Error("Version must not be empty")
	}
}
