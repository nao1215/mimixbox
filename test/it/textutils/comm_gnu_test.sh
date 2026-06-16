Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/comm_gnu
    mkdir -p ${TEST_DIR}
    printf 'apple\nbanana\ncherry\n' > ${TEST_DIR}/a.txt
    printf 'banana\ncherry\ndate\n' > ${TEST_DIR}/b.txt
    printf 'apple\x00banana\x00' > ${TEST_DIR}/za.txt
    printf 'banana\x00date\x00' > ${TEST_DIR}/zb.txt
    printf 'cherry\nbanana\n' > ${TEST_DIR}/unsorted.txt
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/comm_gnu
}

# --output-delimiter replaces the tab column separators with the given string.
TestCommOutputDelimiter() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/comm_gnu
    comm --output-delimiter=, ${TEST_DIR}/a.txt ${TEST_DIR}/b.txt
}

# -z reads and writes NUL-terminated records; pipe through tr to make the NULs
# visible as '#' so shellspec can match a plain string.
TestCommZeroTerminated() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/comm_gnu
    comm -z -1 -2 ${TEST_DIR}/za.txt ${TEST_DIR}/zb.txt | tr '\0' '#'
}

# --check-order fails (exit 1) and prints to stderr when an input is unsorted.
TestCommCheckOrder() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/comm_gnu
    comm --check-order ${TEST_DIR}/unsorted.txt ${TEST_DIR}/b.txt
    echo "rc=$?"
}
