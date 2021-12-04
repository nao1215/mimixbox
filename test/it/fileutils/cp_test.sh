export TEST_DIR=/tmp/mimixbox/it/cp
export DIR_IN_TEST_DIR=/tmp/mimixbox/it/cp/inner
export TEST_DIR2=/tmp/mimixbox/it/cp2
export TEST_FILE=/tmp/mimixbox/it/cp/1.txt
export TEST_FILE2=/tmp/mimixbox/it/cp/2.txt
export TEST_FILE3=/tmp/mimixbox/it/cp/3.txt
export TEST_FILE_INNER=${DIR_IN_TEST_DIR}/inner.txt

Setup() {
    mkdir -p ${TEST_DIR}
    mkdir -p ${TEST_DIR2}
    mkdir -p ${DIR_IN_TEST_DIR}
    touch ${TEST_FILE}
    touch ${TEST_FILE2}
    touch ${TEST_FILE3}
    touch ${TEST_FILE_INNER}
}

Cleanup() {
    rm -rf ${TEST_DIR}
    rm -rf ${TEST_DIR2}
}

TestCopyOneFile() {
    cp ${TEST_FILE} ${TEST_DIR}/cp.txt
    ls ${TEST_DIR}/cp.txt
}

TestCopyOneFileStatus() {
    cp ${TEST_FILE} ${TEST_DIR}/cp.txt
}

TestCopyOndDirWithRecursiveOption() {
    cp -r ${TEST_DIR} ${TEST_DIR2}
    ls ${TEST_DIR2}
    ls ${TEST_DIR2}/cp
}

TestCopyOndDirWithRecursiveOptionStatus() {
    cp -r ${TEST_DIR} ${TEST_DIR2}
}


TestCopySrcAddDistAreSame() {
    cp -r ${TEST_DIR} ${TEST_DIR}
    ls ${TEST_DIR}
}

TestCopySrcAddDistAreSameStatus() {
    cp -r ${TEST_DIR} ${TEST_DIR}
}

TestCopyThreeFileAtSameTime() {
    cp ${TEST_FILE} ${TEST_FILE2} ${TEST_FILE3} ${TEST_DIR2}
    ls ${TEST_DIR2}
}

TestCopyThreeFileAtSameTimeStatus() {
    cp ${TEST_FILE} ${TEST_FILE2} ${TEST_FILE3} ${TEST_DIR2}
}

TestCopyDirctoryWithoutRecursiveOption() {
    cp ${TEST_DIR} ${TEST_DIR2}
    ls ${TEST_DIR2}
}

TestCopyDirctoryWithoutRecursiveOptionStatus() {
    cp ${TEST_DIR} ${TEST_DIR2}
}

TestCopyDirectoryAtRoot() {
    cp -r ${TEST_DIR} /
}