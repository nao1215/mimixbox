Setup() {
    export TEST_DIR=/tmp/mimixbox/it/chmod
    mkdir -p ${TEST_DIR}
    touch ${TEST_DIR}/file.txt
    chmod 600 ${TEST_DIR}/file.txt
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/chmod
}

TestChmodOctal() {
    export TEST_DIR=/tmp/mimixbox/it/chmod
    chmod 644 ${TEST_DIR}/file.txt
    stat -c '%a' ${TEST_DIR}/file.txt
}

TestChmodSymbolic() {
    export TEST_DIR=/tmp/mimixbox/it/chmod
    chmod u+x ${TEST_DIR}/file.txt
    stat -c '%a' ${TEST_DIR}/file.txt
}

TestChmodMissing() {
    chmod 644 /no_such_file
}
