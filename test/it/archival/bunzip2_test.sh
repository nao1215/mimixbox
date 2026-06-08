Setup() {
    export TEST_DIR=/tmp/mimixbox/it/bunzip2
    mkdir -p ${TEST_DIR}
    printf 'bunzip2 payload' | bzip2 -c > ${TEST_DIR}/data.bz2
}
CleanUp() { rm -rf /tmp/mimixbox/it/bunzip2; }

TestBunzip2Stdout() {
    export TEST_DIR=/tmp/mimixbox/it/bunzip2
    bunzip2 -c ${TEST_DIR}/data.bz2
}
