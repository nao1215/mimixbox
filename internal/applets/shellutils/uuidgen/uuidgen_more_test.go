package uuidgen_test

import (
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/uuidgen"
)

// TestSynopsis covers the Synopsis accessor.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	if uuidgen.New().Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestVersionAndVariantBits asserts the generated UUID has the RFC 4122 version
// (4) and variant (8/9/a/b) nibbles in the correct positions.
func TestVersionAndVariantBits(t *testing.T) {
	t.Parallel()
	// Generate several to be sure the variant nibble is always one of 8/9/a/b.
	for i := 0; i < 32; i++ {
		out, errOut, err := run(t)
		if err != nil {
			t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
		}
		id := strings.TrimSpace(out)
		// Layout: xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx
		groups := strings.Split(id, "-")
		if len(groups) != 5 {
			t.Fatalf("uuid %q does not have 5 groups", id)
		}
		if version := groups[2][0]; version != '4' {
			t.Errorf("version nibble = %q, want '4' (uuid=%q)", string(version), id)
		}
		if variant := groups[3][0]; !strings.ContainsRune("89ab", rune(variant)) {
			t.Errorf("variant nibble = %q, want one of 8/9/a/b (uuid=%q)", string(variant), id)
		}
	}
}
