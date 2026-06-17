package timecmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/testutil/fakecmd"
)

// fakeClock makes now return base, then base+elapsed on subsequent calls.
func fakeClock(t *testing.T, elapsed time.Duration) func() {
	t.Helper()
	base := time.Unix(1000, 0)
	seq := []time.Time{base, base.Add(elapsed)}
	i := 0
	orig := now
	now = func() time.Time {
		v := seq[i]
		if i < len(seq)-1 {
			i++
		}
		return v
	}
	return func() { now = orig }
}

// requireCmd installs a repo-local fake of name on PATH so timecmd's exec path
// runs deterministically without depending on host /bin/true, /bin/false, etc.
func requireCmd(t *testing.T, name string) {
	t.Helper()
	fakecmd.UseOnly(t, name)
}

func TestTimeReportsReal(t *testing.T) {
	requireCmd(t, "true")
	defer fakeClock(t, 1500*time.Millisecond)()

	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: errBuf}
	if err := New().Run(context.Background(), io, []string{"true"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(errBuf.String(), "real\t0m1.500s") {
		t.Errorf("stderr = %q, want a real time of 0m1.500s", errBuf.String())
	}
}

func TestTimePosixFormat(t *testing.T) {
	requireCmd(t, "true")
	defer fakeClock(t, 2*time.Second)()

	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: errBuf}
	if err := New().Run(context.Background(), io, []string{"-p", "true"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(errBuf.String(), "real 2.00") {
		t.Errorf("stderr = %q, want 'real 2.00'", errBuf.String())
	}
}

func TestTimePropagatesExitCode(t *testing.T) {
	requireCmd(t, "false")
	defer fakeClock(t, time.Second)()

	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, []string{"false"})
	var ee *command.ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Errorf("err = %v, want exit code 1", err)
	}
}

func TestTimeMissingCommand(t *testing.T) {
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Errorf("missing command should fail")
	}
}
