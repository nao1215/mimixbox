package wc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// twoFiles writes two fixture files and returns their paths. a.txt has 3 lines
// / 3 words / 6 bytes and b.txt has 1 line / 1 word / 2 bytes.
func twoFiles(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("a\nb\nc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return a, b
}

func TestRunTotalModes(t *testing.T) {
	t.Parallel()
	a, b := twoFiles(t)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "only prints just the total without a label",
			args: []string{"--total=only", a, b},
			want: "4 4 8\n",
		},
		{
			name: "never suppresses the total",
			args: []string{"--total=never", a, b},
			want: "3 3 6 " + a + "\n1 1 2 " + b + "\n",
		},
		{
			name: "always prints a total even for one file",
			args: []string{"--total=always", a},
			want: "3 3 6 " + a + "\n3 3 6 total\n",
		},
		{
			name: "auto prints no total for one file",
			args: []string{"--total=auto", a},
			want: "3 3 6 " + a + "\n",
		},
		{
			name: "auto prints a total for two files",
			args: []string{"--total=auto", a, b},
			want: "3 3 6 " + a + "\n1 1 2 " + b + "\n4 4 8 total\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, "", tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRunInvalidTotal(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "--total=bogus")
	if err == nil {
		t.Fatal("expected error for invalid --total")
	}
	if !strings.Contains(errOut, "invalid argument") {
		t.Errorf("stderr = %q, want it to mention an invalid argument", errOut)
	}
}

func TestRunFiles0From(t *testing.T) {
	t.Parallel()
	a, b := twoFiles(t)
	dir := t.TempDir()
	list := filepath.Join(dir, "list.nul")
	// NUL-separated names with a trailing NUL, as GNU wc produces and accepts.
	if err := os.WriteFile(list, []byte(a+"\x00"+b+"\x00"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := run(t, "", "--files0-from="+list)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "3 3 6 " + a + "\n1 1 2 " + b + "\n4 4 8 total\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunFiles0FromStdin(t *testing.T) {
	t.Parallel()
	a, b := twoFiles(t)
	// "-" reads the NUL-separated list from standard input.
	out, _, err := run(t, a+"\x00"+b, "--files0-from=-")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "3 3 6 " + a + "\n1 1 2 " + b + "\n4 4 8 total\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunFiles0FromOnlyTotal(t *testing.T) {
	t.Parallel()
	a, b := twoFiles(t)
	out, _, err := run(t, a+"\x00"+b+"\x00", "--files0-from=-", "--total=only")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "4 4 8\n" {
		t.Errorf("out = %q, want %q", out, "4 4 8\n")
	}
}

func TestRunFiles0FromRejectsOperands(t *testing.T) {
	t.Parallel()
	a, _ := twoFiles(t)
	_, errOut, err := run(t, "", "--files0-from=-", a)
	if err == nil {
		t.Fatal("expected error when operands are combined with --files0-from")
	}
	if !strings.Contains(errOut, "files0-from") {
		t.Errorf("stderr = %q, want it to mention files0-from", errOut)
	}
}
