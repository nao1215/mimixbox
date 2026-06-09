// Package wget implements the wget applet: a non-interactive network downloader
// that fetches each URL operand to a local file (or to standard output), with a
// subset of the common GNU wget options (-O, -q, -P, -c, -T, -t, --user-agent).
package wget

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

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
	output    string
	quiet     bool
	prefix    string
	cont      bool
	timeout   float64
	tries     int
	userAgent string
}

// Run executes wget.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... URL...", stdio.Err).WithHelp(command.Help{
		Description: "Download each URL over HTTP(S). By default the response is saved to a file " +
			"named after the URL in the current directory; -O writes to a specific file, or to " +
			"standard output when given -. -P chooses a destination directory and -c resumes a " +
			"partial download.",
		Examples: []command.Example{
			{Command: "wget https://example.com/a.tar.gz", Explain: "Save to a.tar.gz in the current directory."},
			{Command: "wget -P downloads https://example.com/a.tar.gz", Explain: "Save under the downloads/ directory."},
			{Command: "wget -c https://example.com/big.iso", Explain: "Resume a partially downloaded file."},
			{Command: "wget -O - https://example.com | grep title", Explain: "Stream the body to standard output."},
		},
		ExitStatus: "0  the download succeeded.\n1  a request failed or a file could not be written.",
		Notes: []string{
			"A subset of GNU wget: -O, -q, -P, -c, -T (timeout), -t (tries), and --user-agent.",
		},
	})
	output := fs.StringP("output-document", "O", "", "write documents to FILE (- for standard output)")
	quiet := fs.BoolP("quiet", "q", false, "quiet (no output)")
	prefix := fs.StringP("directory-prefix", "P", "", "save files to DIR")
	cont := fs.BoolP("continue", "c", false, "resume getting a partially downloaded file")
	timeout := fs.Float64P("timeout", "T", 0, "set the network timeout to SECONDS (0 = none)")
	tries := fs.IntP("tries", "t", 1, "set the number of attempts to N")
	userAgent := fs.String("user-agent", "", "identify as STRING to the HTTP server")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	urls := fs.Args()
	if len(urls) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "wget: missing URL")
		return command.SilentFailure()
	}

	opts := options{
		output:    *output,
		quiet:     *quiet,
		prefix:    *prefix,
		cont:      *cont,
		timeout:   *timeout,
		tries:     *tries,
		userAgent: *userAgent,
	}

	client := &http.Client{}
	if opts.timeout > 0 {
		client.Timeout = time.Duration(opts.timeout * float64(time.Second))
	}
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

// retryableErr marks an error as worth retrying (a network failure or a 5xx
// response) so fetch can honor -t/--tries.
type retryableErr struct{ err error }

func (e *retryableErr) Error() string { return e.err.Error() }
func (e *retryableErr) Unwrap() error { return e.err }

func retryable(err error) error { return &retryableErr{err: err} }

func isRetryable(err error) bool {
	var re *retryableErr
	return errors.As(err, &re)
}

// fetch downloads one URL, retrying transient failures up to opts.tries times.
func fetch(ctx context.Context, client *http.Client, stdio command.IO, opts options, rawURL string) error {
	attempts := opts.tries
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		err := fetchOnce(ctx, client, stdio, opts, rawURL)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryable(err) {
			return err
		}
	}
	return lastErr
}

// fetchOnce performs a single download attempt, writing to the file named by -O,
// to standard output when -O is "-", or to a name derived from the URL (placed
// under -P when given). With -c it resumes an existing partial file via a Range
// request.
func fetchOnce(ctx context.Context, client *http.Client, stdio command.IO, opts options, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid URL %q", rawURL)
	}

	dest := destination(opts, parsed)
	toStdout := dest == "-"

	var resumeFrom int64
	if opts.cont && !toStdout {
		if fi, statErr := os.Stat(dest); statErr == nil && fi.Size() > 0 {
			resumeFrom = fi.Size()
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	if opts.userAgent != "" {
		req.Header.Set("User-Agent", opts.userAgent)
	}
	if resumeFrom > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeFrom))
	}

	if !opts.quiet {
		_, _ = fmt.Fprintf(stdio.Err, "wget: connecting to %s ...\n", parsed.Host)
	}

	resp, err := client.Do(req)
	if err != nil {
		return retryable(err)
	}
	defer func() { _ = resp.Body.Close() }()

	appendMode := false
	switch resp.StatusCode {
	case http.StatusOK:
		// Server ignored the range (or none was sent): write from the start.
	case http.StatusPartialContent:
		appendMode = true
	case http.StatusRequestedRangeNotSatisfiable:
		// The local file already has the whole body.
		return nil
	default:
		if resp.StatusCode >= 500 {
			return retryable(fmt.Errorf("server returned %s", resp.Status))
		}
		return fmt.Errorf("server returned %s", resp.Status)
	}

	w, closeFn, err := openDest(stdio, dest, appendMode)
	if err != nil {
		return err
	}

	n, copyErr := io.Copy(w, resp.Body)
	if closeFn != nil {
		if cerr := closeFn(); cerr != nil && copyErr == nil {
			copyErr = cerr
		}
	}
	if copyErr != nil {
		return copyErr
	}

	if !opts.quiet {
		if toStdout {
			_, _ = fmt.Fprintf(stdio.Err, "wget: %d bytes written to standard output\n", n)
		} else {
			_, _ = fmt.Fprintf(stdio.Err, "wget: %q saved [%d]\n", dest, resumeFrom+n)
		}
	}
	return nil
}

// destination returns the local path for a download, honoring -O (exact path)
// and -P (a directory prefix for the URL-derived name).
func destination(opts options, parsed *url.URL) string {
	if opts.output != "" {
		return opts.output
	}
	name := path.Base(parsed.Path)
	if name == "" || name == "." || name == "/" {
		name = "index.html"
	}
	if opts.prefix != "" {
		return filepath.Join(opts.prefix, name)
	}
	return name
}

// openDest returns the writer for dest and an optional close function. When dest
// is "-" the writer is stdio.Out (no close); otherwise a file is created (its
// parent directory is created as needed), opened for append when resuming.
func openDest(stdio command.IO, dest string, appendMode bool) (io.Writer, func() error, error) {
	if dest == "-" {
		return stdio.Out, nil, nil
	}
	if dir := filepath.Dir(dest); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, nil, err
		}
	}
	flags := os.O_WRONLY | os.O_CREATE
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	f, err := os.OpenFile(dest, flags, 0o644) //nolint:gosec // operating on a user-named file is the whole point
	if err != nil {
		return nil, nil, err
	}
	return f, f.Close, nil
}
