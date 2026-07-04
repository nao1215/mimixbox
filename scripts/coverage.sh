#!/bin/sh
# Combine unit-test coverage with self-hosted E2E coverage into a single
# cover.out. Unit tests report line coverage, but they never exercise the real
# multi-call mimixbox binary the way an end user does; the atago E2E specs do
# (every applet on PATH is a symlink to the one instrumented binary). Go 1.20+
# lets us instrument a built binary (`go build -cover`) and collect its runtime
# coverage via GOCOVERDIR, so we can merge "what the unit tests cover" with
# "what real applet invocations cover" and get one honest number.
#
# This is intentionally a separate, heavier target: `make test` / `make e2e`
# stay fast and unchanged. Everything lands under .coverage/ (gitignored) except
# the final cover.out / cover.html, which are the same artifacts `make test`
# already produces so octocov and local tooling need no changes.
#
# Override scenario concurrency with PARALLEL (defaults to 1, matching `make
# e2e`: the kill-family applets signal by name and race under concurrency).
set -eu

cd "$(CDPATH= cd "$(dirname "$0")/.." && pwd)"
root="$(pwd)"
cov="${root}/.coverage"
parallel="${PARALLEL:-1}"

rm -rf "${cov}"
mkdir -p "${cov}/unit" "${cov}/e2e" "${cov}/merged"

# Unit tests read fixtures under /tmp/mimixbox/ut prepared by this script
# (the same step `make pre_ut` runs before `make test`).
echo ">> preparing unit-test fixtures"
"${root}/test/ut/prepareUnitTest.sh"

# 1. Unit-test coverage as raw covdata (GOCOVERDIR form) so it can be merged
#    with the E2E covdata below. -covermode=atomic must match the binary build.
echo ">> unit coverage -> ${cov}/unit"
go test -count=1 -cover -covermode=atomic -coverpkg=./... ./... \
	-args -test.gocoverdir="${cov}/unit"

# 2 + 3. Self-hosted E2E via a coverage-instrumented mimixbox binary. run.sh
#    with COVER set builds `go build -cover` instead of a plain build and
#    installs its applets as symlinks to that one binary; GOCOVERDIR is
#    inherited by every applet child atago spawns, so each writes its own
#    covdata. No spec uses clear_env, so every invocation contributes.
echo ">> e2e coverage -> ${cov}/e2e"
COVER=1 GOCOVERDIR="${cov}/e2e" "${root}/e2e/run.sh" --parallel "${parallel}"

# 4. Merge the raw covdata and render the combined text profile + reports.
echo ">> merging unit + e2e covdata -> cover.out"
go tool covdata merge -i="${cov}/unit,${cov}/e2e" -o="${cov}/merged"
go tool covdata textfmt -i="${cov}/merged" -o="${root}/cover.out"

go tool cover -func=cover.out | tail -n 1
go tool cover -html=cover.out -o cover.html
echo ">> wrote cover.out and cover.html (unit + e2e combined)"
