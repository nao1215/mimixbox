package signal_test

import (
	"testing"

	"github.com/nao1215/mimixbox/internal/signal"
)

func TestNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		spec    string
		want    int
		wantErr bool
	}{
		{"full name", "SIGKILL", 9, false},
		{"short name", "KILL", 9, false},
		{"lower-case", "kill", 9, false},
		{"full term", "SIGTERM", 15, false},
		{"short term", "TERM", 15, false},
		{"number in table", "15", 15, false},
		{"number one", "1", 1, false},
		{"null signal zero", "0", 0, false},
		{"unknown name", "NOPE", 0, true},
		{"unknown number", "999", 0, true},
		{"empty", "", 0, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := signal.Number(tt.spec)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Number(%q) = %d, want error", tt.spec, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Number(%q) unexpected error: %v", tt.spec, err)
			}
			if got != tt.want {
				t.Errorf("Number(%q) = %d, want %d", tt.spec, got, tt.want)
			}
		})
	}
}

func TestNumberLax(t *testing.T) {
	t.Parallel()
	// Unlike Number, an out-of-table number is accepted as-is.
	if n, err := signal.NumberLax("99"); err != nil || n != 99 {
		t.Errorf("NumberLax(99) = %d, %v; want 99, nil", n, err)
	}
	if n, err := signal.NumberLax("KILL"); err != nil || n != 9 {
		t.Errorf("NumberLax(KILL) = %d, %v; want 9, nil", n, err)
	}
	if n, err := signal.NumberLax("SIGTERM"); err != nil || n != 15 {
		t.Errorf("NumberLax(SIGTERM) = %d, %v; want 15, nil", n, err)
	}
	if _, err := signal.NumberLax("BOGUS"); err == nil {
		t.Error("NumberLax(BOGUS) should error")
	}
}

func TestName(t *testing.T) {
	t.Parallel()
	if name, ok := signal.Name(9); !ok || name != "SIGKILL" {
		t.Errorf("Name(9) = %q, %v; want SIGKILL, true", name, ok)
	}
	if _, ok := signal.Name(999); ok {
		t.Error("Name(999) should not be found")
	}
}

// TestRoundTrip asserts the table is non-empty and self-consistent: every entry
// maps name->number and number->name both ways.
func TestRoundTrip(t *testing.T) {
	t.Parallel()
	list := signal.List()
	if len(list) == 0 {
		t.Fatal("signal table is empty")
	}
	seen := make(map[int]bool)
	for _, s := range list {
		if seen[s.Number] {
			t.Errorf("duplicate signal number %d", s.Number)
		}
		seen[s.Number] = true

		n, err := signal.Number(s.Name)
		if err != nil || n != s.Number {
			t.Errorf("Number(%q) = %d, %v; want %d", s.Name, n, err, s.Number)
		}
		name, ok := signal.Name(s.Number)
		if !ok || name != s.Name {
			t.Errorf("Name(%d) = %q, %v; want %q", s.Number, name, ok, s.Name)
		}
	}
}

// TestListIsCopy verifies callers cannot mutate the canonical table.
func TestListIsCopy(t *testing.T) {
	t.Parallel()
	list := signal.List()
	list[0].Name = "MUTATED"
	if again := signal.List(); again[0].Name == "MUTATED" {
		t.Error("List() exposed the underlying table; mutation leaked")
	}
}
