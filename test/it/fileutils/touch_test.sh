export TEST_DIR=/tmp/mimixbox/it/touch

Setup() {
    mkdir -p ${TEST_DIR}
}

Cleanup() {
    rm -rf ${TEST_DIR}
}

TestTouchOneFile() {
    touch ${TEST_DIR}/touch.txt
    ls ${TEST_DIR}/touch.txt
}

TestTouchOneFileStatus() {
    touch ${TEST_DIR}/touch.txt
}

TestTouchThreeFileAtSameTime() {
    touch ${TEST_DIR}/1.txt ${TEST_DIR}/2.txt ${TEST_DIR}/3.txt
    ls  ${TEST_DIR}/1.txt
    ls  ${TEST_DIR}/2.txt
    ls ${TEST_DIR}/3.txt
}

TestTouchThreeFileAtSameTimeStatus() {
    touch ${TEST_DIR}/1.txt ${TEST_DIR}/2.txt ${TEST_DIR}/3.txt
}

TestTouchThreeFileAndNotMakeOneFile() {
    touch ${TEST_DIR}/1.txt /touch/2.txt ${TEST_DIR}/3.txt
    ls  ${TEST_DIR}/1.txt
    ls ${TEST_DIR}/3.txt
}

TestTouchThreeFileAndNotMakeOneFile() {
    touch ${TEST_DIR}/1.txt /touch/2.txt ${TEST_DIR}/3.txt
    ls  ${TEST_DIR}/1.txt
    ls ${TEST_DIR}/3.txt
}

TestTouchThreeFileAndNotMakeOneFileStatus() {
    touch ${TEST_DIR}/1.txt /touch/2.txt ${TEST_DIR}/3.txt
}