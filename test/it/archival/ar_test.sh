Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/ar
    mkdir -p ${TEST_DIR}
    printf 'alpha' > ${TEST_DIR}/a.txt
    printf 'beta'  > ${TEST_DIR}/b.txt
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/ar; }

TestArList() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/ar
    ar rc ${TEST_DIR}/lib.a ${TEST_DIR}/a.txt ${TEST_DIR}/b.txt
    ar t ${TEST_DIR}/lib.a
}
TestArExtract() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/ar
    ar rc ${TEST_DIR}/lib2.a ${TEST_DIR}/a.txt
    mkdir -p ${TEST_DIR}/out && cd ${TEST_DIR}/out
    ar x ${TEST_DIR}/lib2.a
    printf '%s' "$(< ${TEST_DIR}/out/a.txt)"
}
