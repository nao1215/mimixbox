Setup() {
    export TEST_DIR=/tmp/mimixbox/it/realpath
    mkdir -p ${TEST_DIR}
    touch ${TEST_DIR}/file.txt
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/realpath
}

TestRealpathExisting() {
    export TEST_DIR=/tmp/mimixbox/it/realpath
    cd ${TEST_DIR}
    realpath file.txt
}

TestRealpathMissing() {
    realpath -m /tmp/mimixbox/it/realpath/does/not/exist
}

TestRealpathNoOperand() {
    realpath
}
