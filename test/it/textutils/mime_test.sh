# shellcheck shell=sh
# Integration helper for `makemime` / `reformime`. See test/it/README.md.
#
# Exercises a MIME encode -> reparse round trip. Fixtures live under the
# per-run root.

Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mime
    export LANG=C
    mkdir -p "${TEST_DIR}"
    printf 'hello mime\n' > "${TEST_DIR}/part.txt"
}

CleanUp() { rm -rf "${MIMIXBOX_IT_ROOT}/mime"; }

# makemime wraps a file into a single-part MIME message.
TestMakemimeSinglePart() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mime
    makemime "${TEST_DIR}/part.txt"
}

# Round trip: makemime then reformime lists the parts.
TestMakemimeReformimeList() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mime
    makemime "${TEST_DIR}/part.txt" | reformime
}
