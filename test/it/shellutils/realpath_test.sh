Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/realpath
    mkdir -p ${TEST_DIR}
    touch ${TEST_DIR}/file.txt
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/realpath
}

TestRealpathExisting() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/realpath
    cd ${TEST_DIR}
    realpath file.txt
}

TestRealpathMissing() {
    realpath -m ${MIMIXBOX_IT_ROOT}/realpath/does/not/exist
}

TestRealpathNoOperand() {
    realpath
}
