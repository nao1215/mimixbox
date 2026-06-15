Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mknod
    mkdir -p ${TEST_DIR}
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/mknod
}

TestMknodFifo() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mknod
    mknod ${TEST_DIR}/pipe p
    [ -p ${TEST_DIR}/pipe ] && echo fifo
}

TestMknodBadType() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/mknod
    mknod ${TEST_DIR}/x z
}
