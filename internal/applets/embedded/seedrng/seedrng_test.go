package seedrng

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type fakeSeeder struct {
	got    []byte
	credit bool
	called bool
	err    error
}

func (f *fakeSeeder) Credit(seed []byte, credit bool) error {
	f.called = true
	f.got = append([]byte(nil), seed...)
	f.credit = credit
	return f.err
}

func withSeedDir(t *testing.T) {
	t.Helper()
	prev := seedDir
	seedDir = filepath.Join(t.TempDir(), "seedrng")
	t.Cleanup(func() { seedDir = prev })
}

func withSeeder(t *testing.T, s Seeder) {
	t.Helper()
	prev := seeder
	seeder = s
	t.Cleanup(func() { seeder = prev })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return errBuf.String(), err
}

func TestSeedrngFirstRunWritesSeed(t *testing.T) {
	withSeedDir(t)
	fake := &fakeSeeder{}
	withSeeder(t, fake)

	if _, err := run(t); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No existing seed -> seeder not called, but a fresh seed must exist.
	if fake.called {
		t.Error("seeder should not be called on the first run")
	}
	if _, err := os.Stat(filepath.Join(seedDir, seedFile)); err != nil {
		t.Errorf("fresh seed not written: %v", err)
	}
}

func TestSeedrngRestoresAndRefreshes(t *testing.T) {
	withSeedDir(t)
	fake := &fakeSeeder{}
	withSeeder(t, fake)
	if err := os.MkdirAll(seedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	old := []byte("old-seed-bytes")
	if err := os.WriteFile(filepath.Join(seedDir, seedFile), old, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := run(t); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fake.called || string(fake.got) != "old-seed-bytes" {
		t.Errorf("old seed not fed to kernel: called=%v got=%q", fake.called, fake.got)
	}
	if !fake.credit {
		t.Error("expected credit by default")
	}
	// Seed file must have been refreshed to a 256-byte fresh value.
	fresh, err := os.ReadFile(filepath.Join(seedDir, seedFile))
	if err != nil || len(fresh) != 256 {
		t.Errorf("seed not refreshed: len=%d err=%v", len(fresh), err)
	}
}

func TestSeedrngNoCredit(t *testing.T) {
	withSeedDir(t)
	fake := &fakeSeeder{}
	withSeeder(t, fake)
	if err := os.MkdirAll(seedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seedDir, seedFile), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "-n"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.credit {
		t.Error("-n should disable crediting")
	}
}

func TestSeedrngCapabilityError(t *testing.T) {
	withSeedDir(t)
	withSeeder(t, &fakeSeeder{err: errors.New("operation not permitted")})
	if err := os.MkdirAll(seedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seedDir, seedFile), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t); err == nil {
		t.Fatal("expected error when kernel seeding fails")
	}
	// Even on failure, the seed file must have been refreshed first.
	fresh, err := os.ReadFile(filepath.Join(seedDir, seedFile))
	if err != nil || len(fresh) != 256 {
		t.Errorf("seed not refreshed before failure: len=%d err=%v", len(fresh), err)
	}
}
