Setup() { export TEST_DIR=${MIMIXBOX_IT_ROOT}; mkdir -p ${TEST_DIR}; printf 'secret' > ${TEST_DIR}/shred_file; }
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }
TestShredRemove() {
    shred -u ${MIMIXBOX_IT_ROOT}/shred_file
    test ! -e ${MIMIXBOX_IT_ROOT}/shred_file && echo gone
}
