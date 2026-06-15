Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/find
    mkdir -p ${TEST_DIR}/sub
    touch ${TEST_DIR}/a.txt ${TEST_DIR}/sub/b.txt
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/find; }

TestFindName() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/find
    find ${TEST_DIR} -name 'a.txt'
}
TestFindTypeDirCount() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/find
    find ${TEST_DIR} -type d | wc -l | tr -d ' '
}
TestFindUnknown() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/find
    find ${TEST_DIR} -bogus
}
TestFindHelp() {
    find --help
}
TestFindVersion() {
    find --version
}
