Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep
    mkdir -p ${TEST_DIR}
    printf 'apple\nbanana\ncherry\n' > ${TEST_DIR}/fruits.txt
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/grep; }

TestGrepStdin() {
    printf 'one\ntwo\nthree\n' | grep two
}
TestGrepFile() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep
    grep an ${TEST_DIR}/fruits.txt
}
TestGrepCount() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/grep
    grep -c a ${TEST_DIR}/fruits.txt
}
TestGrepNoMatch() {
    printf 'x\n' | grep zzz
}
