Setup() { export TEST_DIR=${MIMIXBOX_IT_ROOT}; mkdir -p ${TEST_DIR}; printf 'hello' > ${TEST_DIR}/stat_file; }
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }
TestStatSize() {
    stat -c '%s' ${MIMIXBOX_IT_ROOT}/stat_file
}
