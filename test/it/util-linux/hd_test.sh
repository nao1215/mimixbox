# shellcheck shell=sh
# Integration helper for `hd` (hexdump in canonical mode). See test/it/README.md.

Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/hd
    export LANG=C
    mkdir -p "${TEST_DIR}"
    printf 'Hello' > "${TEST_DIR}/data.bin"
}

CleanUp() { rm -rf "${MIMIXBOX_IT_ROOT}/hd"; }

# Canonical hex+ASCII dump of a small fixture file.
TestHdFile() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/hd
    hd "${TEST_DIR}/data.bin"
}

# hd reads stdin when no file operand is given.
TestHdStdin() {
    printf 'Hi' | hd
}
