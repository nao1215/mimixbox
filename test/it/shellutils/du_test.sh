Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/du
    mkdir -p ${TEST_DIR}/sub
    printf '%0.s.' $(seq 1 100) > ${TEST_DIR}/a.txt
    printf '%0.s.' $(seq 1 50) > ${TEST_DIR}/sub/b.txt
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/du
}

TestDuBytes() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/du
    du -s -b ${TEST_DIR}
}

TestDuBlocks() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/du
    du -s ${TEST_DIR}
}
