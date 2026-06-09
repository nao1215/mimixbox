# Contributing Guide

Thank you for taking the time to contribute. MimixBox welcomes new applets, options for existing applets, bug fixes, and tests.

## Development Environment

- Go 1.18 or later
- `make`
- `git`
- `shellspec` for the end-to-end tests (`curl -fsSL https://git.io/shellspec | sh -s -- --yes`)
- `golangci-lint` for linting

## Common Commands

```bash
make build      # build the mimixbox binary
make test       # unit tests with coverage (writes cover.out / cover.html)
make test-e2e   # shellspec end-to-end tests against the built binary
make lint       # golangci-lint
```

`make test` (and its `make ut` alias) exits non-zero when any unit test fails, so a failing `go test` fails the local build and the `UnitTest` GitHub Actions workflow. Coverage HTML generation and the temporary-directory cleanup still run afterwards, but they never mask a real test failure.

The end-to-end tests live under `test/it/` and exercise the built binary the way a user does (applet name, flags, stdin, exit codes).

`make test-e2e` is hermetic: it builds MimixBox, stages one symlink per applet in an isolated directory (`test/it/.mbbin`), and runs ShellSpec with that directory first on `PATH`. Every applet therefore resolves to MimixBox, never to a host command of the same name, so the suite can be run in a clean shell without installing MimixBox system-wide. Specs must invoke applets by bare name (e.g. `cat`, `unix2dos`) and must not hardcode an install prefix such as `/usr/local/bin`. `spec/hermetic_spec.sh` guards this contract by asserting that common applets resolve to the MimixBox binary.

## Architecture

Each applet is a small `Command` built on the `internal/command` framework:

```go
type Command interface {
	Name() string
	Synopsis() string
	Run(ctx context.Context, io command.IO, args []string) error
}
```

`command.IO` carries the input and output streams, so an applet never touches `os.Stdin`/`os.Stdout` directly and can be tested entirely in memory. Flags are parsed with `command.NewFlagSet`, a thin wrapper over [spf13/pflag](https://github.com/spf13/pflag) that gives every applet GNU-style parsing (`--long` options, clustered `-abc` short flags, `--` to end options, and operands mixed with options) plus the standard `--help` and `--version`. The goal is for a MimixBox applet to stand in for the system command of the same name, so options should follow GNU coreutils where one exists.

Pure logic that does not need the process (for example text counting or line numbering) lives in a domain package such as `internal/textproc` and is covered by its own unit tests; the applet package stays thin and wires that logic to `command.IO`.

```
project root
├── cmd/mimixbox/main.go          # dispatch
└── internal
    ├── command                   # Command framework, GNU flag helper, exit codes
    ├── textproc                  # pure text logic shared by text applets
    ├── version                   # central version string
    └── applets
        ├── applet.go             # registry: reg(<pkg>.New())
        ├── fileutils
        ├── shellutils
        ├── textutils
        └── ...
```

### Adding or migrating an applet

1. Create (or rewrite) the applet package so its type implements `command.Command`. `cat`, `wc`, `head`, and `basename` are good references.
2. Register it in `internal/applets/applet.go` with `reg(<pkg>.New())`.
3. Add table-driven unit tests for the applet (and for any new `internal/...` logic).
4. Run `make test` and, for behaviour visible from the shell, add a `test/it/` spec.

Older applets still use the legacy `Run() (int, error)` entry point; they are being migrated to the framework one group at a time, and new work should use the `Command` interface.

## Licensing

Contributions must be compatible with the Apache License 2.0. GPLv2-or-later code, and assets such as a Tetris clone, cannot be merged. If you plan to add a command, open a [GitHub Issue](https://github.com/nao1215/mimixbox/issues) first so others do not duplicate the work.

## Pull Request Expectations

- follow GNU coreutils option behaviour for applets that mirror a system command
- add or update tests for new behaviour
- run `make test` (and `make test-e2e` for CLI changes) before opening a PR
- run `make lint` when changing Go code
- record user-facing changes in `CHANGELOG.md` under `[Unreleased]`
