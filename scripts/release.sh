#!/bin/bash -eu
# [Description]
#  Thin wrapper around GoReleaser. `.goreleaser.yml` is the single authoritative
#  release contract: it builds the platform matrix, bundles installer.sh AND
#  libshell.sh into each archive, and produces the deb/rpm/apk packages.
#
#  This script previously reimplemented release packaging in bash, which drifted
#  from .goreleaser.yml (it shipped installer.sh without libshell.sh, so the
#  extracted installer was broken, and it skipped darwin/arm64). That parallel
#  pipeline has been removed; both `make release` and the tagged release CI now
#  go through GoReleaser so the local and tagged outputs are identical.
#
#  Usage:
#    scripts/release.sh            # build a local snapshot into ./dist (no publish)
#    scripts/release.sh --publish  # run a real release (requires a tag + GITHUB_TOKEN)
ROOT_DIR=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
cd "${ROOT_DIR}"

if ! command -v goreleaser >/dev/null 2>&1; then
    echo "ERROR: goreleaser is not installed." >&2
    echo "       Install it from https://goreleaser.com/install/ , e.g.:" >&2
    echo "       go install github.com/goreleaser/goreleaser/v2@latest" >&2
    exit 1
fi

if [ "${1:-}" = "--publish" ]; then
    # Real release: consumes the current git tag and publishes to GitHub.
    exec goreleaser release --clean
fi

# Default: local snapshot build into ./dist with no publishing, so developers
# can inspect the exact artifacts the tagged release would produce.
exec goreleaser release --snapshot --clean
