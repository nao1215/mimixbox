package ipcs

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// header line + one record, matching /proc/sysvipc field order.
	write("msg", "key msqid perms cbytes qnum lspid lrpid uid gid cuid cgid stime rtime ctime\n"+
		"12345 7 600 10 2 100 101 1000 1000 1000 1000 0 0 0\n")
	write("shm", "key shmid perms size cpid lpid nattch uid gid cuid cgid atime dtime ctime rss swap\n"+
		"54321 8 644 4096 200 201 3 1000 1000 1000 1000 0 0 0 0 0\n")
	write("sem", "key semid perms nsems uid gid cuid cgid otime ctime\n"+
		"99 9 666 5 1000 1000 1000 1000 0 0\n")

	orig := procDir
	procDir = dir
	t.Cleanup(func() { procDir = orig })
}

func run(t *testing.T, args ...string) string {
	t.Helper()
	fixture(t)
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, args); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return out.String()
}

func TestAllSections(t *testing.T) {
	out := run(t)
	for _, want := range []string{"Message Queues", "Shared Memory Segments", "Semaphore Arrays"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing section %q", want)
		}
	}
}

func TestMessageQueueRecord(t *testing.T) {
	out := run(t, "-q")
	// 12345 == 0x3039; used-bytes 10; messages 2; perms 600.
	if !strings.Contains(out, "0x00003039") || !strings.Contains(out, "600") {
		t.Errorf("msg record = %q", out)
	}
	if strings.Contains(out, "Shared Memory") {
		t.Errorf("-q should not show shared memory")
	}
}

func TestSharedMemoryRecord(t *testing.T) {
	out := run(t, "-m")
	// 54321 == 0xd431; bytes 4096; nattch 3.
	if !strings.Contains(out, "0x0000d431") || !strings.Contains(out, "4096") || !strings.Contains(out, "644") {
		t.Errorf("shm record = %q", out)
	}
}

func TestSemaphoreRecord(t *testing.T) {
	out := run(t, "-s")
	// 99 == 0x63; nsems 5; perms 666.
	if !strings.Contains(out, "0x00000063") || !strings.Contains(out, "666") {
		t.Errorf("sem record = %q", out)
	}
}

func TestKeyFormat(t *testing.T) {
	t.Parallel()
	if got := key("12345"); got != "0x00003039" {
		t.Errorf("key = %q", got)
	}
	if got := key("0"); got != "0x00000000" {
		t.Errorf("key(0) = %q", got)
	}
}

func TestHelpExitStatus(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help Run error = %v", err)
	}
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing exit status section = %q", out.String())
	}
}
