Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/ln_gnu
    mkdir -p ${TEST_DIR}/a
    mkdir -p ${TEST_DIR}/b
    mkdir -p ${TEST_DIR}/dst
    printf 'content\n' > ${TEST_DIR}/a/target.txt
    printf 'A\n' > ${TEST_DIR}/a.txt
    printf 'B\n' > ${TEST_DIR}/b.txt
}

CleanUp() {
    rm -rf ${MIMIXBOX_IT_ROOT}/ln_gnu
}

# ln -s --relative stores the target relative to the link's own directory.
TestLnRelative() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/ln_gnu
    ln -s --relative ${TEST_DIR}/a/target.txt ${TEST_DIR}/b/link.txt
    readlink ${TEST_DIR}/b/link.txt
}

# ln --target-directory dst a b links each operand into dst (destination-first).
TestLnTargetDirectory() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/ln_gnu
    ln --target-directory ${TEST_DIR}/dst ${TEST_DIR}/a.txt ${TEST_DIR}/b.txt
    ls ${TEST_DIR}/dst
}
