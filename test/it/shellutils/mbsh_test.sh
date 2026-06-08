TestMbshEcho() {
    printf 'echo hello\nexit\n' | mbsh 2>/dev/null
}

TestMbshComment() {
    printf '# a comment\necho ok\nexit\n' | mbsh 2>/dev/null
}

TestMbshLastStatus() {
    printf 'false\necho status=$?\nexit\n' | mbsh 2>/dev/null
}

TestMbshCd() {
    export TEST_DIR=/tmp/mimixbox/it/mbsh
    mkdir -p ${TEST_DIR}
    printf 'cd %s\npwd\nexit\n' "${TEST_DIR}" | mbsh 2>/dev/null
    rm -rf ${TEST_DIR}
}
