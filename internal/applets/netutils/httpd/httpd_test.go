package httpd

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestServesStaticFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	want := "hello mimixbox\n"
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(want), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := &bytes.Buffer{}
	stdio := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}

	var wg sync.WaitGroup
	wg.Add(1)
	var runErr error
	go func() {
		defer wg.Done()
		runErr = New().Run(ctx, stdio, []string{"-f", "-p", "127.0.0.1:0", "-h", dir})
	}()

	addr := waitForAddr(t, out)
	resp, err := http.Get("http://" + addr + "/index.html")
	if err != nil {
		t.Fatalf("GET error: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) != want {
		t.Errorf("body = %q, want %q", string(body), want)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	cancel()
	wg.Wait()
	if runErr != nil {
		t.Errorf("Run returned %v after clean shutdown", runErr)
	}
}

func TestBackgroundModeFails(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), stdio, []string{"-p", "127.0.0.1:0"})
	if err == nil {
		t.Fatal("expected error when -f is omitted, got nil")
	}
}

// waitForAddr polls out until the "serving ... on http://ADDR" line appears.
func waitForAddr(t *testing.T, out *bytes.Buffer) string {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		s := out.String()
		if i := strings.Index(s, "http://"); i >= 0 {
			rest := s[i+len("http://"):]
			if j := strings.IndexByte(rest, '\n'); j >= 0 {
				return strings.TrimSpace(rest[:j])
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not report its address in time")
	return ""
}
