// Package version exposes the MimixBox version and a helper to print it in the
// way GNU coreutils does (program name first, then the project and version).
package version

import (
	"fmt"
	"io"
)

// Version is the MimixBox version. The canonical source is the git tag: both
// `make build` and the GoReleaser release inject it via
// -ldflags "-X github.com/nao1215/mimixbox/internal/version.Version=x.y.z".
// "dev" is the fallback for a plain `go build`/`go install` with no injection.
var Version = "dev"

// Print writes the "--version" line for a single command to w, following the
// GNU coreutils convention, e.g. "cat (mimixbox) 0.36.0".
func Print(w io.Writer, command string) {
	_, _ = fmt.Fprintf(w, "%s (mimixbox) %s\n", command, Version)
}
