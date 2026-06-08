Setup() {
    export TEST_DIR=/tmp/mimixbox/it/patch
    mkdir -p ${TEST_DIR}
    printf 'one\ntwo\nthree\n' > ${TEST_DIR}/f.txt
    {
        printf -- '--- %s\n' "${TEST_DIR}/f.txt"
        printf -- '+++ %s\n' "${TEST_DIR}/f.txt"
        printf '@@ -1,3 +1,3 @@\n one\n-two\n+2\n three\n'
    } > ${TEST_DIR}/p.diff
}
CleanUp() { rm -rf /tmp/mimixbox/it/patch; }

TestPatchApply() {
    export TEST_DIR=/tmp/mimixbox/it/patch
    patch -i ${TEST_DIR}/p.diff >/dev/null
    printf '%s' "$(< ${TEST_DIR}/f.txt)"
}
