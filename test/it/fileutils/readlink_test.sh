Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    mkdir -p ${TEST_DIR}
    printf 'x' > ${TEST_DIR}/rl_target
    ln -sf ${MIMIXBOX_IT_ROOT}/rl_target ${MIMIXBOX_IT_ROOT}/rl_link
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }
TestReadlink() {
    readlink ${MIMIXBOX_IT_ROOT}/rl_link
}
