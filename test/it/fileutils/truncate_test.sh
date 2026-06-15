Setup() { export TEST_DIR=${MIMIXBOX_IT_ROOT}; mkdir -p ${TEST_DIR}; }
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }
TestTruncate() {
    truncate -s 7 ${MIMIXBOX_IT_ROOT}/tr_file
    wc -c < ${MIMIXBOX_IT_ROOT}/tr_file | tr -d ' '
}
