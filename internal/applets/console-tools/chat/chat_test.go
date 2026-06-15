package chat

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestParseScript(t *testing.T) {
	t.Parallel()
	steps, err := ParseScript([]string{"", "ATZ", "OK", "ATDT123", "CONNECT", ""})
	if err != nil {
		t.Fatalf("ParseScript: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("got %d steps", len(steps))
	}
	if steps[0].Expect != "" || steps[0].Send != "ATZ" {
		t.Errorf("step0 = %+v", steps[0])
	}
	if steps[1].Expect != "OK" || steps[1].Send != "ATDT123" {
		t.Errorf("step1 = %+v", steps[1])
	}
	if steps[2].Expect != "CONNECT" || steps[2].Send != "" {
		t.Errorf("step2 = %+v", steps[2])
	}
}

func TestUnescape(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		`a\rb`:  "a\rb",
		`x\ny`:  "x\ny",
		`\\`:    `\`,
		`done\c`: `done\c`,
	}
	for in, want := range cases {
		got, err := unescape(in)
		if err != nil {
			t.Fatalf("unescape(%q): %v", in, err)
		}
		if got != want {
			t.Errorf("unescape(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := unescape(`bad\`); err == nil {
		t.Error("expected error for dangling backslash")
	}
}

// TestConversationOverPipe drives a full expect/send exchange over in-memory
// buffers, the unit-test substitute for a serial link.
func TestConversationOverPipe(t *testing.T) {
	t.Parallel()
	// The "modem" side has already produced "OK\nCONNECT\n".
	link := strings.NewReader("noise OK more CONNECT tail")
	var sent bytes.Buffer

	steps, _ := ParseScript([]string{"", "ATZ", "OK", "ATDT123", "CONNECT", ""})
	if err := Conversation(context.Background(), link, &sent, steps, 0); err != nil {
		t.Fatalf("Conversation: %v", err)
	}
	got := sent.String()
	if !strings.Contains(got, "ATZ\r") || !strings.Contains(got, "ATDT123\r") {
		t.Errorf("sent = %q", got)
	}
}

func TestConversationCSuppressesCR(t *testing.T) {
	t.Parallel()
	link := strings.NewReader("OK")
	var sent bytes.Buffer
	steps, _ := ParseScript([]string{"OK", `done\c`})
	if err := Conversation(context.Background(), link, &sent, steps, 0); err != nil {
		t.Fatalf("Conversation: %v", err)
	}
	if sent.String() != "done" {
		t.Errorf("sent = %q, want %q", sent.String(), "done")
	}
}

func TestConversationExpectNeverSeen(t *testing.T) {
	t.Parallel()
	link := strings.NewReader("nothing useful here")
	var sent bytes.Buffer
	steps, _ := ParseScript([]string{"LOGIN:", "user"})
	if err := Conversation(context.Background(), link, &sent, steps, 0); err == nil {
		t.Error("expected error when expect string never appears")
	}
}

func TestConversationTimeout(t *testing.T) {
	t.Parallel()
	// A reader that blocks forever would need a goroutine; instead use a slow
	// pipe substitute: empty reader returns EOF, so test timeout via a reader
	// that never yields the wanted token but keeps returning bytes.
	link := &repeatReader{b: 'x'}
	var sent bytes.Buffer
	steps, _ := ParseScript([]string{"NEVER", "x"})
	err := Conversation(context.Background(), link, &sent, steps, 20*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error, got %v", err)
	}
}

// repeatReader endlessly yields the same byte, never EOF.
type repeatReader struct{ b byte }

func (r *repeatReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = r.b
	}
	return len(p), nil
}

func TestRunNoScript(t *testing.T) {
	t.Parallel()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Error("expected error when no script given")
	}
}

func TestRunSuccess(t *testing.T) {
	t.Parallel()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader("OK"), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), io, []string{"OK", "GO"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "GO\r") {
		t.Errorf("out = %q", out.String())
	}
}
