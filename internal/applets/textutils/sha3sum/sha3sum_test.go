package sha3sum

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, in string, args ...string) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(in), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, args); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return out.String()
}

func TestDefaultIs256(t *testing.T) {
	t.Parallel()
	// SHA3-256 of "hello\n", verified with openssl dgst -sha3-256.
	const want = "b314e28493eae9dab57ac4f0c6d887bddbbeb810e900d818395ace558e96516d"
	if got := run(t, "hello\n"); !strings.HasPrefix(got, want) {
		t.Errorf("sha3sum = %q, want prefix %q", got, want)
	}
}

func TestAlgorithmSelection(t *testing.T) {
	t.Parallel()
	// SHA3-512 of "hello\n" (openssl).
	const want512 = "ac766ba623301e0ad63c48cb2fc469d1"
	for _, args := range [][]string{{"-a", "512"}, {"-a512"}, {"--algorithm=512"}} {
		got := run(t, "hello\n", args...)
		if !strings.HasPrefix(got, want512) {
			t.Errorf("sha3sum %v = %q, want prefix %q", args, got, want512)
		}
	}
}

func TestExtractAlgo(t *testing.T) {
	t.Parallel()
	bits, rest, err := extractAlgo([]string{"-a", "384", "file.txt"})
	if err != nil || bits != 384 || len(rest) != 1 || rest[0] != "file.txt" {
		t.Errorf("extractAlgo = %d, %v, %v", bits, rest, err)
	}
	bits, _, err = extractAlgo([]string{"file.txt"})
	if err != nil || bits != 256 {
		t.Errorf("default bits = %d, %v; want 256", bits, err)
	}
	if _, _, err := extractAlgo([]string{"-a", "bad"}); err == nil {
		t.Errorf("non-numeric -a should fail")
	}
}

func TestUnsupportedLength(t *testing.T) {
	t.Parallel()
	io := command.IO{In: strings.NewReader("x"), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"-a", "999"}); err == nil {
		t.Errorf("unsupported length should fail")
	}
}
