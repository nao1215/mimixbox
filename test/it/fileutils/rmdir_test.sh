Setup() {
    export TEST_DIR=/tmp/mimixbox/it/rmdir
    mkdir -p ${TEST_DIR}
}

CleanUp() {
    rm -rf /tmp/mimixbox/it/rmdir
}

TestRmdirEmpty() {
    export TEST_DIR=/tmp/mimixbox/it/rmdir
    mkdir -p ${TEST_DIR}/empty
    rmdir ${TEST_DIR}/empty
    test ! -d ${TEST_DIR}/empty && echo "removed"
}

TestRmdirNonEmpty() {
    export TEST_DIR=/tmp/mimixbox/it/rmdir
    mkdir -p ${TEST_DIR}/full
    touch ${TEST_DIR}/full/file.txt
    rmdir ${TEST_DIR}/full
}

TestRmdirParents() {
    export TEST_DIR=/tmp/mimixbox/it/rmdir
    mkdir -p ${TEST_DIR}/a/b/c
    cd ${TEST_DIR}
    rmdir -p a/b/c
    test ! -d ${TEST_DIR}/a && echo "removed"
}

TestRmdirMissingOperand() {
    rmdir
}
