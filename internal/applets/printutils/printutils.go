// Package printutils implements the classic line-printer applets lpr, lpq, and
// lpd against a local spool directory, so they can be exercised end to end
// without a real printer or network daemon.
//
// The spool directory holds one control file per queued job plus the data
// payload. lpr enqueues, lpq lists, and lpd "prints" by draining the queue into
// an output sink (a directory of printed files), all within temp directories
// during tests.
package printutils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// defaultSpool is the spool directory used when -P/SPOOL is not given. Tests set
// the MIMIXBOX_SPOOL via the -S flag instead, so this stays a documented
// default rather than a hidden host write.
const defaultSpool = "/var/spool/mimixbox-lpd"

// job is the metadata stored in a spool control file.
type job struct {
	ID       int       `json:"id"`
	Owner    string    `json:"owner"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Queued   time.Time `json:"queued"`
	DataFile string    `json:"data_file"`
}

// now is the clock; tests override it for deterministic timestamps.
var now = time.Now

// Command is one print applet, distinguished by name.
type Command struct {
	name string
}

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// NewLpr returns the lpr applet (enqueue a print job).
func NewLpr() *Command { return &Command{name: "lpr"} }

// NewLpq returns the lpq applet (list the print queue).
func NewLpq() *Command { return &Command{name: "lpq"} }

// NewLpd returns the lpd applet (drain the print queue).
func NewLpd() *Command { return &Command{name: "lpd"} }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case "lpr":
		return "Queue files for printing into a local spool"
	case "lpq":
		return "Show the local print queue"
	case "lpd":
		return "Drain the local print spool to an output directory"
	}
	return "Line-printer utility"
}

// Run dispatches to the per-applet implementation.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	switch c.name {
	case "lpr":
		return c.runLpr(stdio, args)
	case "lpq":
		return c.runLpq(stdio, args)
	case "lpd":
		return c.runLpd(stdio, args)
	}
	return command.Failuref("%s: unknown print applet", c.name)
}

func (c *Command) runLpr(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-S SPOOL] [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Queue files for printing by copying them into the spool directory (-S SPOOL, " +
			"default " + defaultSpool + ") and writing a control file for each job. With no FILE, or " +
			"'-', the job body is read from standard input. Each job is assigned an increasing numeric " +
			"id; use lpq to list the queue and lpd to drain it. No real printer is contacted.",
		Examples: []command.Example{
			{Command: "lpr -S /tmp/spool document.txt", Explain: "Queue document.txt into /tmp/spool."},
			{Command: "echo hi | lpr -S /tmp/spool", Explain: "Queue standard input."},
		},
		ExitStatus: "0  the job was queued.\n1  the spool or a file could not be accessed.",
	})
	spool := fs.StringP("spool", "S", defaultSpool, "spool directory")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if err := os.MkdirAll(*spool, 0o700); err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
		return command.SilentFailure()
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}
	var failed bool
	for _, f := range files {
		if err := enqueue(*spool, f, stdio.In); err != nil {
			fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

func (c *Command) runLpq(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-S SPOOL]", stdio.Err).WithHelp(command.Help{
		Description: "List the jobs currently queued in the spool directory (-S SPOOL, default " +
			defaultSpool + "), in id order, showing rank, owner, job id, original file name, and size " +
			"in bytes. Reports 'no entries' when the queue is empty.",
		Examples:   []command.Example{{Command: "lpq -S /tmp/spool", Explain: "Show the queue in /tmp/spool."}},
		ExitStatus: "0  success (including an empty queue).\n1  the spool could not be read.",
	})
	spool := fs.StringP("spool", "S", defaultSpool, "spool directory")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	jobs, err := readQueue(*spool)
	if err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
		return command.SilentFailure()
	}
	if len(jobs) == 0 {
		fmt.Fprintln(stdio.Out, "no entries")
		return nil
	}
	tw := tabwriter.NewWriter(stdio.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Rank\tOwner\tJob\tFiles\tSize")
	for i, j := range jobs {
		fmt.Fprintf(tw, "%d\t%s\t%d\t%s\t%d\n", i+1, j.Owner, j.ID, j.Name, j.Size)
	}
	return tw.Flush()
}

func (c *Command) runLpd(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-S SPOOL] -o OUTDIR", stdio.Err).WithHelp(command.Help{
		Description: "Drain the print spool (-S SPOOL, default " + defaultSpool + ") by 'printing' each " +
			"queued job: its data is written into the OUTDIR output directory (-o, created if needed) " +
			"under the job's id and original name, and the job is then removed from the queue. Jobs are " +
			"processed in id order. This is a local, file-based stand-in for the printer daemon; it does " +
			"not open a network socket.",
		Examples: []command.Example{
			{Command: "lpd -S /tmp/spool -o /tmp/printed", Explain: "Print every queued job into /tmp/printed."},
		},
		ExitStatus: "0  the queue was drained.\n1  the spool or output directory could not be accessed.",
		Notes: []string{
			"Network listening is intentionally not implemented; lpd here drains the local spool to a directory.",
		},
	})
	spool := fs.StringP("spool", "S", defaultSpool, "spool directory")
	out := fs.StringP("output", "o", "", "directory to write printed jobs into")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *out == "" {
		fmt.Fprintf(stdio.Err, "%s: no output directory; use -o DIR (network printing is not implemented)\n", c.name)
		return command.SilentFailure()
	}
	jobs, err := readQueue(*spool)
	if err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
		return command.SilentFailure()
	}
	if err := os.MkdirAll(*out, 0o700); err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
		return command.SilentFailure()
	}
	var failed bool
	for _, j := range jobs {
		if err := printJob(*spool, *out, j); err != nil {
			fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
			failed = true
			continue
		}
		fmt.Fprintf(stdio.Out, "printed job %d (%s)\n", j.ID, j.Name)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// enqueue copies one file (or stdin for "-") into the spool and writes its
// control file. The new job id is one past the current maximum.
func enqueue(spool, file string, stdin io.Reader) error {
	id, err := nextID(spool)
	if err != nil {
		return err
	}

	var src io.Reader
	var name string
	if file == "-" {
		src = stdin
		name = "(stdin)"
	} else {
		f, err := os.Open(file) //nolint:gosec // user-named file
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		src = f
		name = filepath.Base(file)
	}

	dataName := fmt.Sprintf("df%04d", id)
	dataPath := filepath.Join(spool, dataName)
	out, err := os.Create(dataPath) //nolint:gosec // path under spool
	if err != nil {
		return err
	}
	size, err := io.Copy(out, src)
	if err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	j := job{
		ID:       id,
		Owner:    owner(),
		Name:     name,
		Size:     size,
		Queued:   now().UTC(),
		DataFile: dataName,
	}
	return writeControl(spool, j)
}

// nextID returns the next free job id by scanning existing control files.
func nextID(spool string) (int, error) {
	jobs, err := readQueue(spool)
	if err != nil {
		return 0, err
	}
	max := 0
	for _, j := range jobs {
		if j.ID > max {
			max = j.ID
		}
	}
	return max + 1, nil
}

// writeControl writes a job's control file as cfNNNN.
func writeControl(spool string, j job) error {
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	cf := filepath.Join(spool, fmt.Sprintf("cf%04d", j.ID))
	return os.WriteFile(cf, append(data, '\n'), 0o600)
}

// readQueue reads all control files in spool, sorted by job id. A missing spool
// directory is an empty queue, not an error.
func readQueue(spool string) ([]job, error) {
	entries, err := os.ReadDir(spool)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var jobs []job
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "cf") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(spool, e.Name())) //nolint:gosec // path under spool
		if err != nil {
			return nil, err
		}
		var j job
		if err := json.Unmarshal(data, &j); err != nil {
			return nil, fmt.Errorf("corrupt control file %s: %w", e.Name(), err)
		}
		jobs = append(jobs, j)
	}
	sort.Slice(jobs, func(i, k int) bool { return jobs[i].ID < jobs[k].ID })
	return jobs, nil
}

// printJob writes a job's data into outDir and removes the job from the spool.
func printJob(spool, outDir string, j job) error {
	src := filepath.Join(spool, j.DataFile)
	in, err := os.Open(src) //nolint:gosec // path under spool
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	safe := strings.ReplaceAll(j.Name, "/", "_")
	dst := filepath.Join(outDir, fmt.Sprintf("%04d-%s", j.ID, safe))
	out, err := os.Create(dst) //nolint:gosec // path under outDir
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	_ = in.Close()
	// Remove control and data files from the spool.
	_ = os.Remove(src)
	return os.Remove(filepath.Join(spool, fmt.Sprintf("cf%04d", j.ID)))
}

// owner returns the current user's login name for the job's Owner field.
func owner() string {
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	if u := os.Getenv("LOGNAME"); u != "" {
		return u
	}
	return "unknown"
}
