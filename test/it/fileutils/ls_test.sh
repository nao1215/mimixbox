Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/ls
    mkdir -p ${TEST_DIR}/sub
    : > ${TEST_DIR}/a.txt
    : > ${TEST_DIR}/b.txt
    : > ${TEST_DIR}/.hidden
}

CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/ls; }

TestLsDefault() { ls ${TEST_DIR}; }
TestLsAll() { ls -a ${TEST_DIR}; }
TestLsClassify() { ls -F ${TEST_DIR}; }
