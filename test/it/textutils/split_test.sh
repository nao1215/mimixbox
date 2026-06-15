Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    mkdir -p ${TEST_DIR}
}
CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}
}
TestSplitLines() {
    printf '1\n2\n3\n' | split -l 2 - ${MIMIXBOX_IT_ROOT}/part-
    cat ${MIMIXBOX_IT_ROOT}/part-aa
}
