Setup() {
    export TEST_DIR=/tmp/mimixbox/it/mbsh
    mkdir -p ${TEST_DIR}
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/mbsh
}

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
    printf 'cd %s\npwd\nexit\n' "${TEST_DIR}" | mbsh 2>/dev/null
}

TestMbshCatConsumesInput() {
    printf 'cat\nhello\nexit\n' | mbsh 2>/dev/null
}

TestMbshNoReparse() {
    err=$(printf 'cat\nhello\nexit\n' | mbsh 2>&1 >/dev/null)
    case "${err}" in
        *"not a mimixbox command"*) echo reparsed ;;
        *) echo ok ;;
    esac
}
