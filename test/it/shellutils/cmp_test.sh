Setup() {
    export TEST_DIR=/tmp/mimixbox/it/cmp
    mkdir -p ${TEST_DIR}
    printf 'abc\n' > ${TEST_DIR}/a.txt
    printf 'abc\n' > ${TEST_DIR}/same.txt
    printf 'abd\n' > ${TEST_DIR}/diff.txt
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/cmp
}

TestCmpEqual() {
    export TEST_DIR=/tmp/mimixbox/it/cmp
    cmp ${TEST_DIR}/a.txt ${TEST_DIR}/same.txt
    echo "rc=$?"
}

TestCmpDiffer() {
    export TEST_DIR=/tmp/mimixbox/it/cmp
    cmp ${TEST_DIR}/a.txt ${TEST_DIR}/diff.txt
}

TestCmpSilent() {
    export TEST_DIR=/tmp/mimixbox/it/cmp
    cmp -s ${TEST_DIR}/a.txt ${TEST_DIR}/diff.txt
    echo "rc=$?"
}
