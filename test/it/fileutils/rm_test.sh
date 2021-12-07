export TEST_DIR=/tmp/mimixbox/it/rm
export DIR_IN_TEST_DIR=${TEST_DIR}/inner
export TEST_FILE1=${TEST_DIR}/1.txt
export TEST_FILE2=${TEST_DIR}/2.txt
export TEST_FILE3=${TEST_DIR}/3.txt
export TEST_FILE_INNER=${DIR_IN_TEST_DIR}/inner.txt

Setup() {
    mkdir -p ${TEST_DIR}
    mkdir -p ${DIR_IN_TEST_DIR}
    touch ${TEST_FILE1}
    touch ${TEST_FILE2}
    touch ${TEST_FILE3}
    touch ${TEST_FILE_INNER}
}

Cleanup() {
    rm -rf ${TEST_DIR}
}

TestRmOneFile() {
    rm ${TEST_FILE1}
    ls ${TEST_DIR}
}

TestRmOneStatus() {
    rm ${TEST_FILE1}
}

TestRmFileWithWildcard() {
    rm *.txt
    ls ${TEST_DIR}
}