Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    mkdir -p ${TEST_DIR}/src
    printf 'log\n' > ${TEST_DIR}/src/a.log
}
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestLogCollect() {
    log-collect -o /tmp/mimixbox/it/out /tmp/mimixbox/it/src >/dev/null
    cat /tmp/mimixbox/it/out/a.log
}
