# shellcheck shell=sh
# Integration helper for `lzop` / `lzopcat` / `unlzop`. See test/it/README.md.
#
# Exercises LZO compression round-trips. Fixtures live under the per-run root.

Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/lzop
    export LANG=C
    mkdir -p "${TEST_DIR}"
    printf 'lzop round trip payload\n' > "${TEST_DIR}/data.txt"
}

CleanUp() { rm -rf "${MIMIXBOX_IT_ROOT}/lzop"; }

# Compress to stdout, decompress via lzopcat: must reproduce the original.
TestLzopRoundTripPipe() {
    printf 'lzo stream\n' | lzop | lzopcat
}

# File-based round trip: lzop creates data.txt.lzo, unlzop restores it.
TestLzopRoundTripFile() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/lzop
    cp "${TEST_DIR}/data.txt" "${TEST_DIR}/rt.txt"
    lzop "${TEST_DIR}/rt.txt"
    unlzop -c "${TEST_DIR}/rt.txt.lzo"
}
