Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/script
    mkdir -p ${TEST_DIR}
}

CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/script; }

TestScriptRecords() {
    script -c 'printf recorded' -T ${TEST_DIR}/timing ${TEST_DIR}/out.txt >/dev/null 2>&1
    grep -c 'Script started' ${TEST_DIR}/out.txt
}

TestScriptReplay() {
    script -c 'printf replayed' -T ${TEST_DIR}/timing ${TEST_DIR}/out.txt >/dev/null 2>&1
    scriptreplay ${TEST_DIR}/timing ${TEST_DIR}/out.txt
}
