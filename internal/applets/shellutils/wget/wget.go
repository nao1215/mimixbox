// Package wget implements the wget applet: a non-interactive network downloader
// that fetches each URL operand to a local file (or to standard output).
package wget

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the wget applet.
type Command struct{}

// New returns a wget command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "wget" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "The non-interactive network downloader" }

type options struct {
	output string
	quiet  bool
}

// Run executes wget.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... URL...", stdio.Err)
	output := fs.StringP("output-document", "O", "", "write documents to FILE (- for standard output)")
	quiet := fs.BoolP("quiet", "q", false, "quiet (no output)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	urls := fs.Args()
	if len(urls) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "wget: missing URL")
		return command.SilentFailure()
	}

	opts := options{output: *output, quiet: *quiet}
	client := &http.Client{}
	return download(ctx, client, stdio, opts, urls)
}

// download fetches every URL with the supplied client. A failure on one URL is
// reported on stderr but does not stop the remaining URLs; the returned error
// only sets the exit code, because its message was already printed.
func download(ctx context.Context, client *http.Client, stdio command.IO, opts options, urls []string) error {
	var firstErr error
	for _, u := range urls {
		if err := fetch(ctx, client, stdio, opts, u); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "wget: %s: %v\n", u, err)
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
		}
	}
	return firstErr
}

// fetch downloads a single URL and writes it to its destination (the file named
// by -O, standard output when -O is "-", or a name derived from the URL path).
func fetch(ctx context.Context, client *http.Client, stdio command.IO, opts options, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid URL %q", rawURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}

	if !opts.quiet {
		_, _ = fmt.Fprintf(stdio.Err, "wget: connecting to %s ...\n", parsed.Host)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	dest := destination(opts, parsed)
	w, toStdout, err := openDest(stdio, dest)
	if err != nil {
		return err
	}

	n, err := io.Copy(w, resp.Body)
	if !toStdout {
		if cerr := w.(io.Closer).Close(); cerr != nil && err == nil {
			err = cerr
		}
	}
	if err != nil {
		return err
	}

	if !opts.quiet {
		if toStdout {
			_, _ = fmt.Fprintf(stdio.Err, "wget: %d bytes written to standard output\n", n)
		} else {
			_, _ = fmt.Fprintf(stdio.Err, "wget: %q saved [%d]\n", dest, n)
		}
	}
	return nil
}

// destination returns the local filename for a download, honoring -O and
// falling back to the base name of the URL path (or "index.html").
func destination(opts options, parsed *url.URL) string {
	if opts.output != "" {
		return opts.output
	}
	name := path.Base(parsed.Path)
	if name == "" || name == "." || name == "/" {
		return "index.html"
	}
	return name
}

// openDest returns the writer for dest. When dest is "-" the writer is
// stdio.Out and toStdout is true; otherwise a file is created and the caller is
// responsible (via the returned io.Closer) for closing it.
func openDest(stdio command.IO, dest string) (io.Writer, bool, error) {
	if dest == "-" {
		return stdio.Out, true, nil
	}
	f, err := os.Create(dest) //nolint:gosec // operating on a user-named file is the whole point
	if err != nil {
		return nil, false, err
	}
	return f, false, nil
}
