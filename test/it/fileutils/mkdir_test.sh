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

TestMkdirSingleStatus() {
    mkdir ${TEST_DIR}/single
}

TestMkdirThreeDirectory() {
    mkdir ${TEST_DIR}/1 ${TEST_DIR}/2 ${TEST_DIR}/3
    ls ${TEST_DIR}
}

TestMkdirThreeDirectoryStatus() {
    mkdir ${TEST_DIR}/1 ${TEST_DIR}/2 ${TEST_DIR}/3
}

TestMkdirParent() {
    mkdir -p ${TEST_DIR}/parents/child
    ls ${TEST_DIR}/parents/
}

TestMkdirParentStatus() {
    mkdir -p ${TEST_DIR}/parents/child
}

TestMkdirFromPipe() {
    echo "${TEST_DIR}/pipe" | xargs mkdir 
    ls ${TEST_DIR}
}

TestMkdirFromPipeStatus() {
    echo "${TEST_DIR}/pipe" | xargs mkdir 
}

TestMkdirNoArg() {
    mkdir
}

TestMkdirNoArgWithParentsOption() {
    mkdir -p
}

TestMkdirThreeDirAndOneIsFail() {
    mkdir ${TEST_DIR}/1 /mkdir/2 ${TEST_DIR}/3
    ls ${TEST_DIR}/
}

TestMkdirThreeDirAndOneIsFailStatus() {
    mkdir ${TEST_DIR}/1 /mkdir/2 ${TEST_DIR}/3
}