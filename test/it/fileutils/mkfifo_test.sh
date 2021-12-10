export TEST_DIR=/tmp/mimixbox/it/mkfifo
export TEST_FILE1=${TEST_DIR}/1
export TEST_FILE2=${TEST_DIR}/2
export TEST_FILE3=${TEST_DIR}/3

Setup() {
    mkdir -p ${TEST_DIR}
}

Cleanup() {
    rm -rf ${TEST_DIR}
}

TestMkfifoOneFifo() {
    mkfifo ${TEST_FILE1}
    ls -al ${TEST_FILE1} | cut -f 1 -d " "
}

TestMkfifoOneFileStatus() {
    mkfifo ${TEST_FILE1}
}

TestMkfifoThreeFifo() {
    mkfifo ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
    ls -al ${TEST_FILE1} | cut -f 1 -d " "
    ls -al ${TEST_FILE2} | cut -f 1 -d " "
    ls -al ${TEST_FILE3} | cut -f 1 -d " "
}

TestMkfifoThreeFifoStatus() {
    mkfifo ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
}

TestMkfifoNoExistPath() {
    mkfifo /no_exist_path/fifo
}

TestMkfifoAlreadyExistSameName() {
    mkfifo ${TEST_FILE1} 
    mkfifo ${TEST_FILE1} 
}

TestMkfifoThreeFileAndCreateOneFileFailed() {
    mkfifo ${TEST_FILE1} /no_exist_path/fifo ${TEST_FILE3}
    ls ${TEST_DIR}
}

TestMkfifoThreeFileAndCreateOneFileFailedStatus() {
    mkfifo ${TEST_FILE1} /no_exist_path/fifo ${TEST_FILE3}
}