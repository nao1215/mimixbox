Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/cmp_gnu
    mkdir -p ${TEST_DIR}
    # Differ at byte 4 ('X' vs 'Y'); identical tail otherwise.
    printf 'abcXdef' > ${TEST_DIR}/a.txt
    printf 'abcYdef' > ${TEST_DIR}/b.txt
    # First three bytes differ; the rest is common.
    printf 'XXXcommon' > ${TEST_DIR}/skip_a.txt
    printf 'YYYcommon' > ${TEST_DIR}/skip_b.txt
    # Asymmetric skip case: skip 1 of file1, 3 of file2, remaining "abZ"/"abQ".
    printf '_abZ' > ${TEST_DIR}/pair_a.txt
    printf '___abQ' > ${TEST_DIR}/pair_b.txt
    # Difference whose byte values matter for --print-bytes ('s' vs 'S').
    printf 'first\nsecond\n' > ${TEST_DIR}/pb_a.txt
    printf 'first\nSECOND\n' > ${TEST_DIR}/pb_b.txt
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/cmp_gnu
}

# -n LIMIT stops before the difference at byte 4, so it reports equality.
TestCmpBytesLimitEqual() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/cmp_gnu
    cmp -n 3 ${TEST_DIR}/a.txt ${TEST_DIR}/b.txt
    echo "rc=$?"
}

# --bytes=4 reaches the differing byte and reports it.
TestCmpBytesLimitDiffer() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/cmp_gnu
    cmp --bytes=4 ${TEST_DIR}/a.txt ${TEST_DIR}/b.txt
}

# -i N skips the first N bytes of both files before comparing.
TestCmpIgnoreInitial() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/cmp_gnu
    cmp -i 3 ${TEST_DIR}/skip_a.txt ${TEST_DIR}/skip_b.txt
    echo "rc=$?"
}

# -i N:M skips N bytes of file1 and M of file2; offsets count from the first
# compared byte.
TestCmpIgnoreInitialPair() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/cmp_gnu
    cmp -i 1:3 ${TEST_DIR}/pair_a.txt ${TEST_DIR}/pair_b.txt
}

# -b adds the octal value and rendered character of each differing byte.
TestCmpPrintBytes() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/cmp_gnu
    cmp -b ${TEST_DIR}/pb_a.txt ${TEST_DIR}/pb_b.txt
}
