Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/chmod
    mkdir -p ${TEST_DIR}
    touch ${TEST_DIR}/file.txt
    chmod 600 ${TEST_DIR}/file.txt
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/chmod
}

TestChmodOctal() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/chmod
    chmod 644 ${TEST_DIR}/file.txt
    stat -c '%a' ${TEST_DIR}/file.txt
}

TestChmodSymbolic() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/chmod
    chmod u+x ${TEST_DIR}/file.txt
    stat -c '%a' ${TEST_DIR}/file.txt
}

TestChmodMissing() {
    chmod 644 /no_such_file
}
