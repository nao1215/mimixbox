Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/patch
    mkdir -p ${TEST_DIR}
    printf 'one\ntwo\nthree\n' > ${TEST_DIR}/f.txt
    {
        printf -- '--- %s\n' "${TEST_DIR}/f.txt"
        printf -- '+++ %s\n' "${TEST_DIR}/f.txt"
        printf '@@ -1,3 +1,3 @@\n one\n-two\n+2\n three\n'
    } > ${TEST_DIR}/p.diff
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/patch; }

TestPatchApply() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/patch
    patch -i ${TEST_DIR}/p.diff >/dev/null
    printf '%s' "$(< ${TEST_DIR}/f.txt)"
}
