Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/bunzip2
    mkdir -p ${TEST_DIR}
    printf 'bunzip2 payload' | bzip2 -c > ${TEST_DIR}/data.bz2
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/bunzip2; }

TestBunzip2Stdout() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/bunzip2
    bunzip2 -c ${TEST_DIR}/data.bz2
}
