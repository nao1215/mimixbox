package cat_test

import (
	"testing"
)

// TestRunShowNonprinting covers GNU cat -v / --show-nonprinting: control
// characters become caret notation (^X), DEL becomes ^?, and high bytes get
// M- notation, while TAB and the trailing newline are left untouched.
func TestRunShowNonprinting(t *testing.T) {
	t.Parallel()
	// 0x01 (control) -> ^A, a literal tab -> stays a tab (only -T converts it),
	// 0x7f (DEL) -> ^?, 0x80 (high byte) -> M-^@, 0xe9 -> M-i.
	stdin := "a\x01\tb\x7f\x80\xe9\n"
	want := "a^A\tb^?M-^@M-i\n"

	tests := []struct {
		name string
		args []string
	}{
		{"short -v", []string{"-v"}},
		{"long --show-nonprinting", []string{"--show-nonprinting"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := runStdin(t, stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != want {
				t.Errorf("out = %q, want %q", out, want)
			}
		})
	}
}

// TestRunShowNonprintingAliasIdentical asserts that -v and --show-nonprinting
// produce byte-identical output on the same input.
func TestRunShowNonprintingAliasIdentical(t *testing.T) {
	t.Parallel()
	stdin := "x\x00\x1f\x7f\x80\xff\ty\n\n"
	short, _, err := runStdin(t, stdin, "-v")
	if err != nil {
		t.Fatalf("-v error = %v", err)
	}
	long, _, err := runStdin(t, stdin, "--show-nonprinting")
	if err != nil {
		t.Fatalf("--show-nonprinting error = %v", err)
	}
	if short != long {
		t.Errorf("-v = %q, --show-nonprinting = %q; want identical", short, long)
	}
}

// TestRunShowAll covers GNU cat -A / --show-all, which is equivalent to -vET:
// non-printing characters via -v notation, TAB shown as ^I (-T), and a $ at
// each line end (-E).
func TestRunShowAll(t *testing.T) {
	t.Parallel()
	// Fixture: a tab, a blank line, and non-printable bytes 0x01, 0x7f, 0x80.
	stdin := "a\tb\x01\n\n\x7f\x80\n"
	// -v: 0x01->^A, 0x7f->^?, 0x80->M-^@; -T: tab->^I; -E: $ at each end.
	want := "a^Ib^A$\n$\n^?M-^@$\n"

	tests := []struct {
		name string
		args []string
	}{
		{"short -A", []string{"-A"}},
		{"long --show-all", []string{"--show-all"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := runStdin(t, stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != want {
				t.Errorf("out = %q, want %q", out, want)
			}
		})
	}
}

// TestRunShowAllEqualsVET verifies that -A is exactly equivalent to -vET.
func TestRunShowAllEqualsVET(t *testing.T) {
	t.Parallel()
	stdin := "a\tb\x01\n\n\x7f\x80\xe9\n"
	a, _, err := runStdin(t, stdin, "-A")
	if err != nil {
		t.Fatalf("-A error = %v", err)
	}
	vet, _, err := runStdin(t, stdin, "-v", "-E", "-T")
	if err != nil {
		t.Fatalf("-vET error = %v", err)
	}
	if a != vet {
		t.Errorf("-A = %q, -vET = %q; want identical", a, vet)
	}
}

// TestRunShowNonprintingLeavesTab confirms that -v alone never converts TAB
// (only -T does) and leaves the trailing newline intact.
func TestRunShowNonprintingLeavesTab(t *testing.T) {
	t.Parallel()
	out, _, err := runStdin(t, "a\tb\n", "-v")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "a\tb\n" {
		t.Errorf("out = %q, want %q", out, "a\tb\n")
	}
}
