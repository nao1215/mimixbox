Setup() {
    export TEST_DIR=/tmp/mimixbox/it/ls
    mkdir -p ${TEST_DIR}/sub
    : > ${TEST_DIR}/a.txt
    : > ${TEST_DIR}/b.txt
    : > ${TEST_DIR}/.hidden
}

CleanUp() { rm -rf /tmp/mimixbox/it/ls; }

TestLsDefault() { ls ${TEST_DIR}; }
TestLsAll() { ls -a ${TEST_DIR}; }
TestLsClassify() { ls -F ${TEST_DIR}; }
