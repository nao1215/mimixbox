Setup() {
    export TEST_DIR=/tmp/mimixbox/it/ln
    mkdir -p ${TEST_DIR}
    printf 'content\n' > ${TEST_DIR}/target.txt
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/ln
}

TestLnHard() {
    export TEST_DIR=/tmp/mimixbox/it/ln
    ln ${TEST_DIR}/target.txt ${TEST_DIR}/hardlink.txt
    cat ${TEST_DIR}/hardlink.txt
}

TestLnSymbolic() {
    export TEST_DIR=/tmp/mimixbox/it/ln
    ln -s ${TEST_DIR}/target.txt ${TEST_DIR}/symlink.txt
    test -L ${TEST_DIR}/symlink.txt && echo "is symlink"
}

TestLnNoOperand() {
    ln
}
