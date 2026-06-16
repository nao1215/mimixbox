package whoami_test

import (
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/whoami"
)

func TestSynopsis(t *testing.T) {
	t.Parallel()
	c := whoami.New()
	if c.Name() != "whoami" {
		t.Errorf("Name() = %q, want whoami", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// The user.Current() failure path in Run is not exercised: it can only fail when
// the current UID has no passwd entry, which is not reproducible deterministically
// in a unit test without a real broken environment.
