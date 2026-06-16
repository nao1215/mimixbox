package printutils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

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

// spool is the local print-queue backend shared by the lpr/lpq/lpd front-ends.
// It owns the queue storage policy — control files (cfNNNN), data payloads
// (dfNNNN), id allocation, listing, and draining — so the CLI fronts only parse
// flags and render output. All operations stay local; no network is used.
type spool struct {
	dir string
}

// openSpool returns a spool rooted at dir. It does not create the directory:
// enqueue creates it on demand and list treats a missing directory as empty.
func openSpool(dir string) *spool { return &spool{dir: dir} }

// init ensures the spool directory exists, ready for enqueueing.
func (s *spool) init() error { return os.MkdirAll(s.dir, 0o700) }

// enqueue copies one file (or stdin for "-") into the spool and writes its
// control file. The new job id is one past the current maximum.
func (s *spool) enqueue(file string, stdin io.Reader) error {
	id, err := s.nextID()
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
	dataPath := filepath.Join(s.dir, dataName)
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
	return s.writeControl(j)
}

// nextID returns the next free job id by scanning existing control files.
func (s *spool) nextID() (int, error) {
	jobs, err := s.list()
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
func (s *spool) writeControl(j job) error {
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	cf := filepath.Join(s.dir, fmt.Sprintf("cf%04d", j.ID))
	return os.WriteFile(cf, append(data, '\n'), 0o600)
}

// list reads all control files in the spool, sorted by job id. A missing spool
// directory is an empty queue, not an error.
func (s *spool) list() ([]job, error) {
	entries, err := os.ReadDir(s.dir)
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
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name())) //nolint:gosec // path under spool
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

// print writes a job's data into outDir and removes the job from the spool. The
// control and data files are removed only after the payload is written, so a
// failed print leaves the job queued.
func (s *spool) print(outDir string, j job) error {
	src := filepath.Join(s.dir, j.DataFile)
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
	return os.Remove(filepath.Join(s.dir, fmt.Sprintf("cf%04d", j.ID)))
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
