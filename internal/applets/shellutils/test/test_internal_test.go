package testcmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHasPerm exercises the pure permission predicate across read, write and
// execute bits for owner, group and other, including the empty-permission case.
func TestHasPerm(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mode os.FileMode
		op   string
		want bool
	}{
		{"owner read", 0o400, "-r", true},
		{"group read", 0o040, "-r", true},
		{"other read", 0o004, "-r", true},
		{"no read", 0o222, "-r", false},
		{"owner write", 0o200, "-w", true},
		{"no write", 0o555, "-w", false},
		{"owner exec", 0o100, "-x", true},
		{"other exec", 0o001, "-x", true},
		{"no exec", 0o666, "-x", false},
		{"unknown op", 0o777, "-q", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := hasPerm(tt.mode, tt.op); got != tt.want {
				t.Errorf("hasPerm(%o, %q) = %v, want %v", tt.mode, tt.op, got, tt.want)
			}
		})
	}
}

// TestEvalFileTestExisting drives the file-test branches that depend on the
// real mode bits of files created on disk.
func TestEvalFileTestExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	regular := filepath.Join(dir, "reg")
	if err := os.WriteFile(regular, []byte("content"), 0o644); err != nil {
		t.Fatalf("setup regular: %v", err)
	}
	empty := filepath.Join(dir, "empty")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatalf("setup empty: %v", err)
	}
	readonly := filepath.Join(dir, "ro")
	if err := os.WriteFile(readonly, []byte("x"), 0o444); err != nil {
		t.Fatalf("setup readonly: %v", err)
	}
	exec := filepath.Join(dir, "exec")
	if err := os.WriteFile(exec, []byte("x"), 0o755); err != nil {
		t.Fatalf("setup exec: %v", err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(regular, link); err != nil {
		t.Fatalf("setup symlink: %v", err)
	}
	missing := filepath.Join(dir, "missing")

	tests := []struct {
		name    string
		op      string
		path    string
		want    bool
		wantErr bool
	}{
		{"-e regular", "-e", regular, true, false},
		{"-e missing", "-e", missing, false, false},
		{"-f regular", "-f", regular, true, false},
		{"-f dir", "-f", dir, false, false},
		{"-d dir", "-d", dir, true, false},
		{"-d regular", "-d", regular, false, false},
		{"-s nonempty", "-s", regular, true, false},
		{"-s empty", "-s", empty, false, false},
		{"-b regular", "-b", regular, false, false},
		{"-c regular", "-c", regular, false, false},
		{"-p regular", "-p", regular, false, false},
		{"-S regular", "-S", regular, false, false},
		{"-L symlink", "-L", link, true, false},
		{"-h regular", "-h", regular, false, false},
		{"-L missing", "-L", missing, false, false},
		{"-r readonly", "-r", readonly, true, false},
		{"-w readonly", "-w", readonly, false, false},
		{"-w regular", "-w", regular, true, false},
		{"-x exec", "-x", exec, true, false},
		{"-x regular", "-x", regular, false, false},
		{"-r missing", "-r", missing, false, false},
		{"unknown op", "-Q", regular, false, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := evalFileTest(tt.op, tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("evalFileTest(%q, ...) err = %v, wantErr %v", tt.op, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("evalFileTest(%q, ...) = %v, want %v", tt.op, got, tt.want)
			}
		})
	}
}

// TestSynopsis covers the one-line description accessor.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	if New().Synopsis() == "" {
		t.Error("Synopsis() = empty")
	}
}
