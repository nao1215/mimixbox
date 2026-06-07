Setup() {
    export TEST_DIR=/tmp/mimixbox/it/serial
    mkdir -p ${TEST_DIR}
    touch ${TEST_DIR}/apple.txt ${TEST_DIR}/banana.txt ${TEST_DIR}/cherry.txt
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/serial
}

TestSerialRename() {
    export TEST_DIR=/tmp/mimixbox/it/serial
    serial ${TEST_DIR} > /dev/null
    ls ${TEST_DIR}
}

TestSerialDryRun() {
    export TEST_DIR=/tmp/mimixbox/it/serial
    serial -d ${TEST_DIR} > /dev/null
    ls ${TEST_DIR}
}

TestSerialNoOperand() {
    serial
}
