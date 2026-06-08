Setup() {
    export TEST_DIR=/tmp/mimixbox/it/gunzip
    mkdir -p ${TEST_DIR}
    printf 'gunzip payload' > ${TEST_DIR}/data
    gzip ${TEST_DIR}/data
}
CleanUp() { rm -rf /tmp/mimixbox/it/gunzip; }

TestGunzipStdout() {
    export TEST_DIR=/tmp/mimixbox/it/gunzip
    gunzip -c ${TEST_DIR}/data.gz
}
