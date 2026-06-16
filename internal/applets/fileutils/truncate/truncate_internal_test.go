package truncate

import "testing"

// TestParseBytes covers the plain, suffixed, and invalid byte-count forms.
func TestParseBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		spec    string
		want    int64
		wantErr bool
	}{
		{"plain", "100", 100, false},
		{"zero", "0", 0, false},
		{"K", "1K", 1024, false},
		{"KB", "2KB", 2 * 1024, false},
		{"M", "1M", 1024 * 1024, false},
		{"MB", "3MB", 3 * 1024 * 1024, false},
		{"G", "1G", 1024 * 1024 * 1024, false},
		{"GB", "2GB", 2 * 1024 * 1024 * 1024, false},
		{"non-numeric", "abc", 0, true},
		{"negative", "-5", 0, true},
		{"empty number with suffix", "K", 0, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseBytes(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseBytes(%q) err = %v, wantErr %v", tt.spec, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseBytes(%q) = %d, want %d", tt.spec, got, tt.want)
			}
		})
	}
}

// TestResolveSize covers absolute, relative (+/-) and the negative-clamp paths.
func TestResolveSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		spec    string
		cur     int64
		want    int64
		wantErr bool
	}{
		{"absolute", "500", 100, 500, false},
		{"absolute with suffix", "1K", 0, 1024, false},
		{"grow", "+50", 100, 150, false},
		{"grow suffix", "+1K", 0, 1024, false},
		{"shrink", "-30", 100, 70, false},
		{"shrink below zero clamps", "-200", 100, 0, false},
		{"invalid", "+xyz", 0, 0, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := resolveSize(tt.spec, tt.cur)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveSize(%q, %d) err = %v, wantErr %v", tt.spec, tt.cur, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("resolveSize(%q, %d) = %d, want %d", tt.spec, tt.cur, got, tt.want)
			}
		})
	}
}

func TestSynopsis(t *testing.T) {
	t.Parallel()
	if New().Synopsis() == "" {
		t.Error("Synopsis() = empty")
	}
}
