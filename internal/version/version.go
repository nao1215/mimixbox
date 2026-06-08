// Package version exposes the MimixBox version and a helper to print it in the
// way GNU coreutils does (program name first, then the project and version).
package version

import (
	"fmt"
	"io"
)

// Version is the MimixBox version. It is over_written at build time with
// -ldflags "-X github.com/nao1215/mimixbox/internal/version.Version=x.y.z".
var Version = "0.35.0"

// Print writes the "--version" line for a single command to w, following the
// GNU coreutils convention, e.g. "cat (mimixbox) 0.35.0".
func Print(w io.Writer, command string) {
	_, _ = fmt.Fprintf(w, "%s (mimixbox) %s\n", command, Version)
}
