# Testing

Each technique below is here because it catches a kind of bug the others miss. The commands run from the root of a checkout.

## Unit tests

Every applet has tests next to it that run it through `command.Execute` with an in-memory standard input and output. `make test` runs them with coverage; they read fixtures that `make pre_ut` writes. CI: `unit_test.yml`, on the supported Go versions and with the race detector.

## End-to-end scenarios

`e2e/atago/` holds about 1,500 [atago](https://github.com/nao1215/atago) scenarios that run the built binary through the applet names on an isolated `PATH`. `make e2e` runs them one at a time, because some scenarios signal processes by name (`killall`, `pkill`) and would stop a scenario running beside them. CI: `integration_test.yml`.

## Differential tests against GNU coreutils

`internal/applets/differential_test.go` generates text with [rapid](https://github.com/flyingmutant/rapid) (separators, tabs, digits, case, repeated and empty lines, a last line with or without its newline) and runs each applet and GNU coreutils on it in the C locale. The output and the exit status must be the same. rapid shrinks a failure to its smallest input and prints the seed. The test finds GNU under its plain name or under `gnu<name>` (Ubuntu 26.04 installs uutils under the plain names) and skips the cases where GNU is absent.

`cut -c`, `expand`, `unexpand`, `cat -A` and `od -c` are compared on ASCII input only: GNU counts bytes in the C locale where mimixbox counts characters, which is a choice about multibyte text rather than a bug.

When the test was added it found eight differences, each fixed with a test of its own: `base64` and `base32` wrote a newline for empty input and after `-w 0`; `fold` added a newline the input did not have and counted tabs, backspaces and carriage returns as one column; `nl` left a last line without its newline; `cat -E` marked a line end where there was no newline; `sort -n` misread keys such as `7-` and `+5`; `sort -k` rebuilt the key without its blanks; and `unexpand` replaced a lone space before a tab stop with a tab and expanded tabs it was not asked to convert.

`internal/applets/archival/lzopcomp/compat_test.go` compares the lzop applet with upstream `lzop` in both directions when it is installed.

## Fuzzing

Six fuzz targets cover the lzop decoder and its round trip, `chmod` modes, `cut` ranges and the escapes of `echo` and `printf`. CI: `fuzz.yml` searches with each target every night; `go test` runs their corpora on every pull request.

## Benchmarks

`bench/` measures start-up and a few applets end to end with [himorime](https://github.com/nao1215/himorime). On a pull request `bench.yml` builds the base and the head and measures them in alternating rounds on one runner, and fails when the head is slower or uses more memory beyond the tolerance of the suite. See bench/README.md.

## Mutation testing

[gremlins](https://github.com/go-gremlins/gremlins) v0.6.0 changes one expression at a time and checks that a test fails. A pull request mutates the lines it changes; every Sunday the packages below are mutated whole, one job each (`mutation.yml`). Run one package with `make pre_ut` and then `gremlins unleash --workers 2 ./internal/applets/textutils/tr`.

The score is killed / (killed + lived); mutants no test reaches are counted apart as not covered. The values below are the first run, on 2026-11-19 (main 5320885), and a later run is compared with them rather than with a fixed threshold, since some surviving mutants change nothing observable. A high score does not mean an applet agrees with GNU: `sort` scored 100% on the day the differential test found two of its bugs.

<!-- mutation-scores:start -->
| Package | Killed | Lived | Not covered | Score |
|---|---|---|---|---|
| `internal/hashsum` | 23 | 0 | 0 | 100.00% |
| `internal/lib` | 124 | 0 | 6 | 100.00% |
| `internal/textproc` | 71 | 11 | 0 | 86.59% |
| `internal/applets/archival/lzopcomp` | 57 | 3 | 5 | 95.00% |
| `internal/applets/shellutils/base64` | 18 | 3 | 15 | 85.71% |
| `internal/applets/shellutils/chmod` | 42 | 1 | 0 | 97.67% |
| `internal/applets/shellutils/cut` | 57 | 13 | 3 | 81.43% |
| `internal/applets/shellutils/echo` | 47 | 7 | 0 | 87.04% |
| `internal/applets/shellutils/printf` | 73 | 15 | 0 | 82.95% |
| `internal/applets/shellutils/sort` | 132 | 0 | 32 | 100.00% |
| `internal/applets/shellutils/uniq` | 25 | 1 | 0 | 96.15% |
| `internal/applets/textutils/cat` | 20 | 0 | 10 | 100.00% |
| `internal/applets/textutils/fold` | 15 | 2 | 0 | 88.24% |
| `internal/applets/textutils/nl` | 9 | 1 | 0 | 90.00% |
| `internal/applets/textutils/tr` | 99 | 28 | 4 | 77.95% |
| `internal/applets/textutils/unexpand` | 36 | 9 | 0 | 80.00% |
| `internal/applets/textutils/wc` | 25 | 3 | 0 | 89.29% |
| all | 873 | 97 | 75 | 90.00% |
<!-- mutation-scores:end -->

The applets not listed are wiring around a library call or need root, a terminal or the network; their changed lines are mutated on every pull request.
