Setup() {
    export TEST_DIR=/tmp/mimixbox/it/tee
    mkdir -p ${TEST_DIR}
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/tee
}

TestTeeStdoutAndFile() {
    export TEST_DIR=/tmp/mimixbox/it/tee
    printf 'hello\n' | tee ${TEST_DIR}/out.txt
}

TestTeeFileContents() {
    export TEST_DIR=/tmp/mimixbox/it/tee
    printf 'hello\n' | tee ${TEST_DIR}/out.txt > /dev/null
    cat ${TEST_DIR}/out.txt
}

TestTeeAppend() {
    export TEST_DIR=/tmp/mimixbox/it/tee
    printf 'one\n' | tee ${TEST_DIR}/log.txt > /dev/null
    printf 'two\n' | tee -a ${TEST_DIR}/log.txt > /dev/null
    cat ${TEST_DIR}/log.txt
}
