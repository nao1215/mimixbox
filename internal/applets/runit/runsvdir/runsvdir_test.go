package runsvdir

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// stub records which services were started and returns a race-safe snapshot
// accessor.
func stub(t *testing.T) func() []string {
	t.Helper()
	var mu sync.Mutex
	var started []string
	or, os2 := startRunsvFn, rescanInterval
	rescanInterval = time.Millisecond
	startRunsvFn = func(ctx context.Context, dir string, _ command.IO) {
		mu.Lock()
		started = append(started, filepath.Base(dir))
		mu.Unlock()
		<-ctx.Done()
	}
	t.Cleanup(func() { startRunsvFn, rescanInterval = or, os2 })
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string{}, started...)
	}
}

func TestStartsEachService(t *testing.T) {
	snapshot := stub(t)
	dir := t.TempDir()
	for _, name := range []string{"web", "db", ".hidden"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	_ = os.WriteFile(filepath.Join(dir, "afile"), nil, 0o644) // not a dir

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		io := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		done <- New().Run(ctx, io, []string{dir})
	}()

	deadline := time.Now().Add(2 * time.Second)
	for len(snapshot()) < 2 {
		if time.Now().After(deadline) {
			cancel()
			t.Fatalf("started only %v", snapshot())
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	<-done

	got := snapshot()
	sort.Strings(got)
	if len(got) != 2 || got[0] != "db" || got[1] != "web" {
		t.Errorf("started = %v, want [db web] (hidden and files skipped)", got)
	}
}

func TestErrors(t *testing.T) {
	stub(t)
	io := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Errorf("missing directory should fail")
	}
	if err := New().Run(context.Background(), io, []string{"/no/such/dir"}); err == nil {
		t.Errorf("a nonexistent directory should fail")
	}
}
