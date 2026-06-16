// Command genapplets regenerates the applet registry in
// internal/applets/applet_registry_gen.go from the applet packages under
// internal/applets/**. It scans every package for exported nullary constructors
// (New, NewHalt, ...) that return *Command, and emits an import block plus an
// init() that registers each constructed command under its own Name(). This
// removes the hand-maintained import block and map that previously had to be
// edited by hand for every new applet.
//
// Run it with `make generate` (or `go generate ./...`).
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	modulePath = "github.com/nao1215/mimixbox"
	appletsRel = "internal/applets"
	outputRel  = "internal/applets/applet_registry_gen.go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "genapplets:", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := moduleRoot()
	if err != nil {
		return err
	}
	pkgs, err := scan(filepath.Join(root, appletsRel))
	if err != nil {
		return err
	}
	src, err := render(pkgs)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, outputRel), src, 0o644) //nolint:gosec // generated source is world-readable
}

// moduleRoot walks up from the working directory until it finds go.mod, so the
// generator works regardless of which directory `go generate` runs it from.
func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

// pkg is one applet package and the constructors it exposes.
type pkg struct {
	importPath string
	alias      string
	ctors      []string
	// subsystem is the applet family (top-level directory under
	// internal/applets, e.g. "textutils", "procps").
	subsystem string
	// stability is the maturity classification emitted for every command in
	// this package; see stabilityFor.
	stability string
}

// subsystemFor returns the applet family for a package path relative to
// internal/applets: the first path component (e.g. "textutils/cat" ->
// "textutils", "printutils" -> "printutils").
func subsystemFor(rel string) string {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	return parts[0]
}

// gatedSubsystems are families dominated by privileged/destructive operations
// (mounts, raw block devices, login/security/power surfaces), so their applets
// default to the "gated" stability. Everything else defaults to "stable".
var gatedSubsystems = map[string]bool{
	"loginutils":    true,
	"securityutils": true,
	"pmutils":       true,
	"util-linux":    true,
	"embedded":      true,
	"runit":         true,
	"console-tools": true,
}

// stabilityFor returns the default stability for a subsystem. Privileged
// families are "gated"; everything else (coreutils-style text/shell/file
// utilities, games, jokes, ...) is "stable".
func stabilityFor(subsystem string) string {
	if gatedSubsystems[subsystem] {
		return "gated"
	}
	return "stable"
}

// stabilityConst maps a stability value to the exported applets.Stability
// constant used in the generated source.
func stabilityConst(stability string) string {
	switch stability {
	case "gated":
		return "StabilityGated"
	case "partial":
		return "StabilityPartial"
	case "experimental":
		return "StabilityExperimental"
	default:
		return "StabilityStable"
	}
}

// scan walks the applet tree and returns, sorted by import path, every package
// that exposes one or more exported nullary constructors returning *Command.
func scan(appletsDir string) ([]pkg, error) {
	byDir := map[string][]string{}
	fset := token.NewFileSet()

	err := filepath.WalkDir(appletsDir, func(path string, d fs.DirEntry, err error) error {
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
			if !ok || fn.Recv != nil {
				continue
			}
			if !strings.HasPrefix(fn.Name.Name, "New") || !isNullary(fn) || !returnsCommandPtr(fn) {
				continue
			}
			dir := filepath.Dir(path)
			byDir[dir] = append(byDir[dir], fn.Name.Name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	pkgs := make([]pkg, 0, len(byDir))
	for dir, ctors := range byDir {
		rel, rerr := filepath.Rel(appletsDir, dir)
		if rerr != nil {
			return nil, rerr
		}
		sort.Strings(ctors)
		sub := subsystemFor(rel)
		pkgs = append(pkgs, pkg{
			importPath: modulePath + "/" + appletsRel + "/" + filepath.ToSlash(rel),
			alias:      aliasFor(rel),
			ctors:      ctors,
			subsystem:  sub,
			stability:  stabilityFor(sub),
		})
	}
	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].importPath < pkgs[j].importPath })
	return pkgs, nil
}

func isNullary(fn *ast.FuncDecl) bool {
	return fn.Type.Params == nil || len(fn.Type.Params.List) == 0
}

// returnsCommandPtr reports whether fn returns exactly one *Command.
func returnsCommandPtr(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}
	star, ok := fn.Type.Results.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	id, ok := star.X.(*ast.Ident)
	return ok && id.Name == "Command"
}

// aliasFor turns a package's path relative to internal/applets into a unique,
// collision-free import alias (e.g. "debianutils/add-shell" -> "ap_debianutils_add_shell").
func aliasFor(rel string) string {
	r := strings.NewReplacer("/", "_", "-", "_", ".", "_")
	return "ap_" + r.Replace(filepath.ToSlash(rel))
}

// render produces the gofmt-formatted source of the generated registry file.
func render(pkgs []pkg) ([]byte, error) {
	total := 0
	for _, p := range pkgs {
		total += len(p.ctors)
	}

	var b bytes.Buffer
	b.WriteString("// Code generated by \"go generate\"; DO NOT EDIT.\n")
	b.WriteString("//\n")
	b.WriteString("// Run `make generate` (or `go generate ./...`) after adding or removing an\n")
	b.WriteString("// applet package to regenerate the registry.\n")
	b.WriteString("package applets\n\n")

	b.WriteString("import (\n")
	for _, p := range pkgs {
		fmt.Fprintf(&b, "\t%s %q\n", p.alias, p.importPath)
	}
	b.WriteString(")\n\n")

	b.WriteString("// init populates the applet table. Each command is registered under its own\n")
	b.WriteString("// Name(), so the key can never drift from the command it dispatches to.\n")
	b.WriteString("func init() {\n")
	fmt.Fprintf(&b, "\tApplets = make(map[string]Applet, %d)\n", total)
	for _, p := range pkgs {
		stab := stabilityConst(p.stability)
		for _, c := range p.ctors {
			fmt.Fprintf(&b, "\tregister(%s.%s(), %q, %s)\n", p.alias, c, p.subsystem, stab)
		}
	}
	b.WriteString("}\n")

	return format.Source(b.Bytes())
}
