Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    mkdir -p ${TEST_DIR}
    printf 'data' > ${TEST_DIR}/link_src
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }
TestLink() {
    link ${MIMIXBOX_IT_ROOT}/link_src ${MIMIXBOX_IT_ROOT}/link_dst
    cat ${MIMIXBOX_IT_ROOT}/link_dst
}
