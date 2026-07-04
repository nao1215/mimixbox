#!/usr/bin/env bash
#
# run.sh builds the mimixbox multi-call binary from this checkout,
# `--full-install`s every applet into an isolated bin dir, puts that dir FIRST
# on PATH, and runs the atago end-to-end suite (e2e/atago/**/*.atago.yaml)
# against the real applets.
#
# The test DEFINITIONS are atago YAML — this script is only the environment
# bootstrap (a plain shell program, not a test framework).
#
# Because the applet bin dir is first on PATH, bare commands like `cat`, `seq`,
# `split`, `sort`, `tr` resolve to mimixbox's own applets — never to a host
# command of the same name — so the suite runs hermetically in a clean shell
# without installing MimixBox system-wide. Each atago scenario runs in its own
# isolated temp workdir.
#
# Usage: e2e/run.sh [atago args...]        (e.g. e2e/run.sh --parallel 8)
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"

if ! command -v atago >/dev/null 2>&1; then
	echo "e2e: atago is not installed. Install it from https://github.com/nao1215/atago" >&2
	echo "e2e: e.g. 'go install github.com/nao1215/atago@latest' (CI uses nao1215/setup-atago)" >&2
	exit 127
fi

TMP="$(mktemp -d "${TMPDIR:-/tmp}/mimixbox-e2e.XXXXXX")"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT
mkdir -p "$TMP/applets"

# COVER=1 (used by scripts/coverage.sh) builds a coverage-instrumented binary
# so every applet invocation writes covdata to GOCOVERDIR. The default path is
# byte-for-byte the plain build, keeping `make e2e` fast and unchanged.
if [ -n "${COVER:-}" ]; then
	: "${GOCOVERDIR:?COVER=1 requires GOCOVERDIR to be set}"
	echo "e2e: building the coverage-instrumented mimixbox binary..."
	(cd "$REPO_ROOT" && go build -cover -covermode=atomic -coverpkg=./... -trimpath -o "$TMP/applets/mimixbox" ./cmd/mimixbox)
else
	echo "e2e: building the mimixbox binary..."
	(cd "$REPO_ROOT" && go build -trimpath -o "$TMP/applets/mimixbox" ./cmd/mimixbox)
fi

echo "e2e: installing applets via --full-install..."
"$TMP/applets/mimixbox" --full-install "$TMP/applets" >/dev/null

# Put the applet dir FIRST on PATH so bare command names resolve to mimixbox's
# own applets (cat, seq, split, sort, ...).
export PATH="$TMP/applets:$PATH"

echo "e2e: applets installed: $(find "$TMP/applets" -mindepth 1 | wc -l)"
# Extra args (e.g. --parallel 8) go before the path so the flag parser sees them.
atago run "$@" "$SCRIPT_DIR/atago"
