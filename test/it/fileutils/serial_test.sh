Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/serial
    mkdir -p ${TEST_DIR}
    touch ${TEST_DIR}/apple.txt ${TEST_DIR}/banana.txt ${TEST_DIR}/cherry.txt
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/serial
}

TestSerialRename() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/serial
    serial ${TEST_DIR} > /dev/null
    ls ${TEST_DIR}
}

TestSerialDryRun() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/serial
    serial -d ${TEST_DIR} > /dev/null
    ls ${TEST_DIR}
}

TestSerialNoOperand() {
    serial
}
