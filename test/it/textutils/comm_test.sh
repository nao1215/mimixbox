Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    mkdir -p ${TEST_DIR}
    printf 'apple\nbanana\n' > ${TEST_DIR}/comm_a.txt
    printf 'banana\ncherry\n' > ${TEST_DIR}/comm_b.txt
}
CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}
}
TestCommBoth() {
    comm -1 -2 ${MIMIXBOX_IT_ROOT}/comm_a.txt ${MIMIXBOX_IT_ROOT}/comm_b.txt
}
