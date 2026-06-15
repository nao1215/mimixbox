Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/cmp
    mkdir -p ${TEST_DIR}
    printf 'abc\n' > ${TEST_DIR}/a.txt
    printf 'abc\n' > ${TEST_DIR}/same.txt
    printf 'abd\n' > ${TEST_DIR}/diff.txt
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/cmp
}

TestCmpEqual() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/cmp
    cmp ${TEST_DIR}/a.txt ${TEST_DIR}/same.txt
    echo "rc=$?"
}

TestCmpDiffer() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/cmp
    cmp ${TEST_DIR}/a.txt ${TEST_DIR}/diff.txt
}

TestCmpSilent() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/cmp
    cmp -s ${TEST_DIR}/a.txt ${TEST_DIR}/diff.txt
    echo "rc=$?"
}
