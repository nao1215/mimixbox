export TEST_DIR=/tmp/mimixbox/it/mkdir

Setup() {
    mkdir -p ${TEST_DIR}
}

Cleanup() {
    rm -rf ${TEST_DIR}
}

TestMkdirSingle() {
    mkdir ${TEST_DIR}/single
    ls ${TEST_DIR}
}

TestMkdirParent() {
    mkdir -p ${TEST_DIR}/parents/child
    ls ${TEST_DIR}/parents/
}

TestMkdirFromPipe() {
    echo "${TEST_DIR}/pipe" | xargs mkdir 
    ls ${TEST_DIR}
}

TestMkdirNoArg() {
    mkdir
}

TestMkdirNoArgWithParentsOption() {
    mkdir -p
}