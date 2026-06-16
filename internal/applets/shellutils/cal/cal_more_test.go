package cal_test

import (
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/cal"
)

// TestCenter covers both branches of center(): a short string is padded on the
// left only (extra odd space goes right, so the result is not right-padded),
// and a string at least as wide as the field is returned unchanged.
func TestCenter(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		width int
		want  string
	}{
		{"even padding", "ab", 6, "  ab"},
		{"odd padding floors left", "abc", 6, " abc"},
		{"equal width unchanged", "abcdef", 6, "abcdef"},
		{"wider than field unchanged", "toolongtitle", 6, "toolongtitle"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := cal.Center(tt.s, tt.width); got != tt.want {
				t.Errorf("Center(%q,%d) = %q, want %q", tt.s, tt.width, got, tt.want)
			}
		})
	}
}
