#!/bin/bash -eu
# [Description]
#  Smoke check that every release archive bundles all files the shipped
#  installer.sh needs to install MimixBox and its license documentation.
#
#  The release contract (see .goreleaser.yml archives, scripts/installer.sh,
#  and the Makefile "licenses" target) requires each extracted archive to
#  contain:
#    - mimixbox            (the binary, built by GoReleaser)
#    - LICENSE             (MimixBox's own license)
#    - licenses/           (dependency-license output from go-licenses)
#    - installer.sh        (self-contained installer)
#    - libshell.sh         (helpers sourced by installer.sh)
#
# [Usage]
#  scripts/check_release_archive.sh DIR
#
#  DIR is either:
#    - a GoReleaser "dist" directory containing *.tar.gz archives, or
#    - a single already-extracted archive directory.

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <dist-dir-or-extracted-archive-dir>" >&2
    exit 2
fi

TARGET_DIR="$1"
if [ ! -d "${TARGET_DIR}" ]; then
    echo "ERROR: '${TARGET_DIR}' is not a directory" >&2
    exit 2
fi

# Files that must be present at the root of every extracted archive.
# "licenses" is a directory; the rest are regular files.
REQUIRED_FILES="mimixbox LICENSE installer.sh libshell.sh"
REQUIRED_DIRS="licenses"

# checkExtractedDir verifies a single extracted archive root contains every
# installer-required asset. Returns non-zero on the first missing asset.
checkExtractedDir() {
    local root="$1"
    local missing=0 f d

    for f in ${REQUIRED_FILES}; do
        if [ ! -f "${root}/${f}" ]; then
            echo "  MISSING file: ${f}" >&2
            missing=1
        fi
    done
    for d in ${REQUIRED_DIRS}; do
        if [ ! -d "${root}/${d}" ]; then
            echo "  MISSING dir:  ${d}/" >&2
            missing=1
        elif [ -z "$(ls -A "${root}/${d}")" ]; then
            echo "  EMPTY dir:    ${d}/" >&2
            missing=1
        fi
    done
    return ${missing}
}

# resolveArchiveRoot prints the directory that actually holds the archive
# contents. GoReleaser archives use wrap_in_directory, so an extracted tarball
# contains a single top-level directory; descend into it when present.
resolveArchiveRoot() {
    local dir="$1"
    local entries
    entries=$(ls -A "${dir}")
    if [ "$(echo "${entries}" | wc -l)" -eq 1 ] && [ -d "${dir}/${entries}" ]; then
        echo "${dir}/${entries}"
    else
        echo "${dir}"
    fi
}

failures=0
checked=0

# Collect any tarball archives directly under TARGET_DIR.
shopt -s nullglob
archives=("${TARGET_DIR}"/*.tar.gz)
shopt -u nullglob

if [ "${#archives[@]}" -gt 0 ]; then
    workdir=$(mktemp -d "${TMPDIR:-/tmp}/mbb-archive-check.XXXXXX")
    trap 'rm -rf "${workdir}"' EXIT INT TERM
    for archive in "${archives[@]}"; do
        echo "Checking archive: $(basename "${archive}")"
        dest="${workdir}/$(basename "${archive}" .tar.gz)"
        mkdir -p "${dest}"
        tar -xzf "${archive}" -C "${dest}"
        root=$(resolveArchiveRoot "${dest}")
        if ! checkExtractedDir "${root}"; then
            echo "  FAIL: $(basename "${archive}") is missing required files" >&2
            failures=$((failures + 1))
        else
            echo "  OK"
        fi
        checked=$((checked + 1))
    done
else
    # No archives: treat TARGET_DIR as an already-extracted archive root.
    echo "No *.tar.gz found under ${TARGET_DIR}; treating it as an extracted archive."
    root=$(resolveArchiveRoot "${TARGET_DIR}")
    echo "Checking extracted dir: ${root}"
    if ! checkExtractedDir "${root}"; then
        echo "  FAIL: ${root} is missing required files" >&2
        failures=$((failures + 1))
    else
        echo "  OK"
    fi
    checked=$((checked + 1))
fi

if [ "${checked}" -eq 0 ]; then
    echo "ERROR: nothing checked under ${TARGET_DIR}" >&2
    exit 2
fi

if [ "${failures}" -ne 0 ]; then
    echo "Release archive smoke check FAILED (${failures} archive(s) incomplete)." >&2
    exit 1
fi

echo "Release archive smoke check passed (${checked} archive(s))."
