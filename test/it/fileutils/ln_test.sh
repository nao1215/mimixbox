Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/ln
    mkdir -p ${TEST_DIR}
    printf 'content\n' > ${TEST_DIR}/target.txt
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/ln
}

TestLnHard() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/ln
    ln ${TEST_DIR}/target.txt ${TEST_DIR}/hardlink.txt
    cat ${TEST_DIR}/hardlink.txt
}

TestLnSymbolic() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/ln
    ln -s ${TEST_DIR}/target.txt ${TEST_DIR}/symlink.txt
    test -L ${TEST_DIR}/symlink.txt && echo "is symlink"
}

TestLnNoOperand() {
    ln
}
