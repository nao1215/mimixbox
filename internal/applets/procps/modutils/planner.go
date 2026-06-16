package modutils

import (
	"debug/elf"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// This file is the shared module-name/file parsing and planning backend used by
// every modutils applet (insmod, rmmod, modprobe, depmod). The per-command files
// hold only the CLI surface and delegate the hermetic validation, ELF/.modinfo
// inspection, and dependency planning here. Behavior and output are unchanged;
// the logic is merely centralized so the applets share one implementation.

// gatedError is the standard documented refusal for a privileged kernel
// mutation. detail describes the validated plan so the message states what was
// checked before the capability gate stopped it.
func gatedError(name, detail string) error {
	return command.Failuref(
		"%s: %s validated successfully, but inserting/removing kernel modules requires CAP_SYS_MODULE; "+
			"this privileged step is intentionally not implemented in the hermetic build", name, detail)
}

// validateModuleFile checks that path is an ELF object carrying a .modinfo
// section, the minimum that identifies a kernel module.
func validateModuleFile(path string) error {
	f, err := os.Open(path) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	ef, err := elf.NewFile(f)
	if err != nil {
		return fmt.Errorf("not an ELF object: %w", err)
	}
	if ef.Section(".modinfo") == nil {
		return fmt.Errorf("missing .modinfo section (not a kernel module?)")
	}
	return nil
}

// validateModuleName checks that name is a bare module name, not a path.
func validateModuleName(name string) error {
	if name == "" {
		return fmt.Errorf("empty module name")
	}
	if strings.ContainsAny(name, "/") || strings.HasSuffix(name, ".ko") {
		return fmt.Errorf("%q is not a bare module name (use the module name, e.g. 'loop')", name)
	}
	return nil
}

// validateModuleNames validates every name in names, returning the first error.
func validateModuleNames(names []string) error {
	for _, n := range names {
		if err := validateModuleName(n); err != nil {
			return err
		}
	}
	return nil
}

// dep is one module and the modules it depends on, as parsed from .modinfo.
type dep struct {
	path string
	deps []string
}

// buildDeps scans dir for *.ko files and extracts each module's "depends"
// field, producing a dependency map. It is fully hermetic.
func buildDeps(dir string) ([]dep, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var deps []dep
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".ko") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		field, err := modinfoField(full, "depends")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		var names []string
		if field != "" {
			for _, n := range strings.Split(field, ",") {
				if n != "" {
					names = append(names, n)
				}
			}
		}
		deps = append(deps, dep{path: e.Name(), deps: names})
	}
	return deps, nil
}

// modinfoField returns one field value from a .ko file's .modinfo section.
func modinfoField(path, key string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // user-named file
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	ef, err := elf.NewFile(f)
	if err != nil {
		return "", fmt.Errorf("not an ELF object: %w", err)
	}
	sec := ef.Section(".modinfo")
	if sec == nil {
		return "", fmt.Errorf("no .modinfo section")
	}
	data, err := sec.Data()
	if err != nil {
		return "", err
	}
	for _, rec := range strings.Split(string(data), "\x00") {
		if i := strings.IndexByte(rec, '='); i >= 0 && rec[:i] == key {
			return rec[i+1:], nil
		}
	}
	return "", nil
}
