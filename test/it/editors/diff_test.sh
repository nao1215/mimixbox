Setup() {
    export TEST_DIR=/tmp/mimixbox/it/diff
    mkdir -p ${TEST_DIR}
    printf 'one\ntwo\nthree\n' > ${TEST_DIR}/a
    printf 'one\n2\nthree\n'   > ${TEST_DIR}/b
    printf 'one\ntwo\nthree\n' > ${TEST_DIR}/c
}
CleanUp() { rm -rf /tmp/mimixbox/it/diff; }

TestDiffNormal() {
    export TEST_DIR=/tmp/mimixbox/it/diff
    diff ${TEST_DIR}/a ${TEST_DIR}/b
}
TestDiffSame() {
    export TEST_DIR=/tmp/mimixbox/it/diff
    diff ${TEST_DIR}/a ${TEST_DIR}/c
}
TestDiffBrief() {
    export TEST_DIR=/tmp/mimixbox/it/diff
    diff -q ${TEST_DIR}/a ${TEST_DIR}/b
}
