Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    mkdir -p ${TEST_DIR}
    printf 'x' > ${TEST_DIR}/unlink.txt
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }
TestUnlink() {
    unlink ${MIMIXBOX_IT_ROOT}/unlink.txt
    test ! -e ${MIMIXBOX_IT_ROOT}/unlink.txt && echo gone
}
