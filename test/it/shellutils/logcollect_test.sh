Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    mkdir -p ${TEST_DIR}/src
    printf 'log\n' > ${TEST_DIR}/src/a.log
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }
TestLogCollect() {
    log-collect -o ${MIMIXBOX_IT_ROOT}/out ${MIMIXBOX_IT_ROOT}/src >/dev/null
    cat ${MIMIXBOX_IT_ROOT}/out/a.log
}
