Setup() {
    export TEST_DIR=/tmp/mimixbox/it/script
    mkdir -p ${TEST_DIR}
}

CleanUp() { rm -rf /tmp/mimixbox/it/script; }

TestScriptRecords() {
    script -c 'printf recorded' -T ${TEST_DIR}/timing ${TEST_DIR}/out.txt >/dev/null 2>&1
    grep -c 'Script started' ${TEST_DIR}/out.txt
}

TestScriptReplay() {
    script -c 'printf replayed' -T ${TEST_DIR}/timing ${TEST_DIR}/out.txt >/dev/null 2>&1
    scriptreplay ${TEST_DIR}/timing ${TEST_DIR}/out.txt
}
