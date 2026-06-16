package sync_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/sync"
	"github.com/nao1215/mimixbox/internal/command"
)

// TestRunUnknownFlag drives Run's flag-parse failure branch (proceed == false /
// err != nil), which the happy-path tests do not reach.
func TestRunUnknownFlag(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	if err := sync.New().Run(context.Background(), io, []string{"--definitely-not-a-flag"}); err == nil {
		t.Fatal("expected an error for an unknown flag")
	}
}
