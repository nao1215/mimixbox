Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/tar
    mkdir -p ${TEST_DIR}/src
    printf 'alpha' > ${TEST_DIR}/src/a.txt
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/tar; }

TestTarRoundTrip() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/tar
    tar -c -f ${TEST_DIR}/out.tar -C ${TEST_DIR} src
    mkdir -p ${TEST_DIR}/dest
    tar -x -f ${TEST_DIR}/out.tar -C ${TEST_DIR}/dest
    printf '%s' "$(< ${TEST_DIR}/dest/src/a.txt)"
}
TestTarList() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/tar
    tar -c -f ${TEST_DIR}/list.tar -C ${TEST_DIR} src
    tar -t -f ${TEST_DIR}/list.tar
}
