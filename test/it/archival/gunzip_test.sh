Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/gunzip
    mkdir -p ${TEST_DIR}
    printf 'gunzip payload' > ${TEST_DIR}/data
    gzip ${TEST_DIR}/data
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/gunzip; }

TestGunzipStdout() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/gunzip
    gunzip -c ${TEST_DIR}/data.gz
}
