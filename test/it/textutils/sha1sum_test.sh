export TEST_DIR=/tmp/mimixbox/it/sha1sum
export TEST_FILE1=${TEST_DIR}/1.txt
export TEST_FILE2=${TEST_DIR}/2.txt
export TEST_FILE3=${TEST_DIR}/3.txt
export CHECK_SUM_FILE=${TEST_DIR}/checksum.txt

Setup() {
    mkdir -p ${TEST_DIR}
    echo "Dungeon of Regalias" > ${TEST_FILE1} 
    echo "DEMONION" > ${TEST_FILE2}
    echo "Dungeon Crusadearz" > ${TEST_FILE3}

    echo "9dc2936d38932f9ffc6738cb677e4a8722116070  /tmp/mimixbox/it/sha1sum/1.txt" > ${CHECK_SUM_FILE}
    echo "317e30648976d62fae4662fe4435e6568648e8a7  /tmp/mimixbox/it/sha1sum/2.txt" >> ${CHECK_SUM_FILE}
    echo "d4e9619d949de0c0182a09757346ad22e80114b3  /tmp/mimixbox/it/sha1sum/3.txt" >> ${CHECK_SUM_FILE}
}

Cleanup() {
    rm -rf ${TEST_DIR}
}

TestMd5sumOneFile() {
    sha1sum ${TEST_FILE1}
}

TestMd5sumOneDirectory() {
    sha1sum ${TEST_DIR}
}

TestMd5sumNotExistFile() {
    sha1sum /not_exist_file
}

TestMd5sumThreeFiles() {
    sha1sum ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
}

TestMd5sumWithCheckOption() {
    sha1sum -c ${CHECK_SUM_FILE}
}

TestMd5sumDataFromPipe() {
    echo "test" | sha1sum 
}

TestMd5sumFileAndDataFromPipeAtSameTime() {
    echo "test" | sha1sum ${TEST_FILE1}
}
