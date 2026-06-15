// mimixbox/internal/applets/applet_test.go
//
// # Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package applets

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/pidof"
)

// TestPidofIsRegistered guards against the regression where pidof had a real
// implementation and unit tests in the tree but was never wired into the applet
// registry. Without registration, "mimixbox pidof", "--list" and
// "--full-install" silently omit it, and the ShellSpec suite ends up exercising
// the host pidof instead of MimixBox's own. See GitHub issue #265.
func TestPidofIsRegistered(t *testing.T) {
	t.Parallel()

	name := pidof.New().Name()
	if !HasApplet(name) {
		t.Fatalf("applet %q is implemented but not registered in Applets", name)
	}

	if got, want := Applets[name].Desc, pidof.New().Synopsis(); got != want {
		t.Errorf("registered description drifted from the command synopsis:\n got: %q\nwant: %q", got, want)
	}
}

// TestEveryConstructorIsRegistered asserts that every applet package under
// internal/applets/** exposing a New*() *Command constructor is present in the
// generated registry. A forgotten applet shows up here as a count mismatch,
// independently of the generator. The test runs with the package directory as
// the working directory, so "." is the internal/applets tree.
func TestEveryConstructorIsRegistered(t *testing.T) {
	t.Parallel()

	got := countAppletConstructors(t, ".")
	if got != len(Applets) {
		t.Errorf("source has %d New*() *Command constructors but the registry has %d applets; run `make generate`", got, len(Applets))
	}
}

// TestRegistryKeyMatchesName asserts every map key equals its command's Name(),
// so a key can never dispatch to a differently-named command.
func TestRegistryKeyMatchesName(t *testing.T) {
	t.Parallel()

	for key, applet := range Applets {
		if name := applet.Cmd.Name(); name != key {
			t.Errorf("registry key %q maps to a command whose Name() is %q", key, name)
		}
	}
}

// TestListAppletsTo asserts the "name - description" table is written to the
// injected writer (not os.Stdout) and that every registered applet appears with
// its synopsis. This is the writer-injected replacement for the removed
// process-global ListApplets wrapper (issue #492).
func TestListAppletsTo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ListAppletsTo(&out)
	got := out.String()

	if len(Applets) == 0 {
		t.Fatal("registry is empty; cannot validate the listing")
	}
	for name, applet := range Applets {
		line := name + " - " + applet.Desc
		if !strings.Contains(got, line) {
			t.Errorf("ListAppletsTo output is missing %q", line)
		}
	}
	if want := len(Applets); strings.Count(got, "\n") != want {
		t.Errorf("ListAppletsTo wrote %d lines, want %d", strings.Count(got, "\n"), want)
	}
}

// TestShowAppletsBySpaceSeparatedTo asserts the space-separated listing goes to
// the injected writer, wraps at the 60-column boundary, and contains every
// applet name exactly once. Replacement for the removed process-global wrapper.
func TestShowAppletsBySpaceSeparatedTo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	ShowAppletsBySpaceSeparatedTo(&out)
	got := out.String()

	if !strings.HasSuffix(got, "\n") {
		t.Errorf("output should end with a newline, got %q", got[max(0, len(got)-10):])
	}
	fields := strings.Fields(got)
	if len(fields) != len(Applets) {
		t.Errorf("output has %d names, want %d", len(fields), len(Applets))
	}
	for _, name := range SortApplet() {
		if !strings.Contains(got, name) {
			t.Errorf("output is missing applet %q", name)
		}
	}
	for _, line := range strings.Split(strings.TrimRight(got, "\n"), "\n") {
		if len(line) > 60 {
			t.Errorf("line exceeds the 60-column wrap: %q (%d)", line, len(line))
		}
	}
}

// countAppletConstructors parses the Go sources under dir and counts exported
// nullary constructors (New, NewHalt, ...) that return *Command.
func countAppletConstructors(t *testing.T, dir string) int {
	t.Helper()
	fset := token.NewFileSet()
	count := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "New") {
				continue
			}
			if fn.Type.Params != nil && len(fn.Type.Params.List) != 0 {
				continue
			}
			if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
				continue
			}
			star, ok := fn.Type.Results.List[0].Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			if id, ok := star.X.(*ast.Ident); ok && id.Name == "Command" {
				count++
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return count
}
