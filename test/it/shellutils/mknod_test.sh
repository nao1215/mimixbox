Setup() {
    export TEST_DIR=/tmp/mimixbox/it/mknod
    mkdir -p ${TEST_DIR}
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/mknod
}

TestMknodFifo() {
    export TEST_DIR=/tmp/mimixbox/it/mknod
    mknod ${TEST_DIR}/pipe p
    [ -p ${TEST_DIR}/pipe ] && echo fifo
}

TestMknodBadType() {
    export TEST_DIR=/tmp/mimixbox/it/mknod
    mknod ${TEST_DIR}/x z
}
