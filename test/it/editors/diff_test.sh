Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/diff
    mkdir -p ${TEST_DIR}
    printf 'one\ntwo\nthree\n' > ${TEST_DIR}/a
    printf 'one\n2\nthree\n'   > ${TEST_DIR}/b
    printf 'one\ntwo\nthree\n' > ${TEST_DIR}/c
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/diff; }

TestDiffNormal() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/diff
    diff ${TEST_DIR}/a ${TEST_DIR}/b
}
TestDiffSame() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/diff
    diff ${TEST_DIR}/a ${TEST_DIR}/c
}
TestDiffBrief() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/diff
    diff -q ${TEST_DIR}/a ${TEST_DIR}/b
}
