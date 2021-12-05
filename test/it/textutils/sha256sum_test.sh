export TEST_DIR=/tmp/mimixbox/it/sha256sum
export TEST_FILE1=${TEST_DIR}/1.txt
export TEST_FILE2=${TEST_DIR}/2.txt
export TEST_FILE3=${TEST_DIR}/3.txt
export CHECK_SUM_FILE=${TEST_DIR}/checksum.txt

Setup() {
    mkdir -p ${TEST_DIR}
    echo "Dungeon of Regalias" > ${TEST_FILE1} 
    echo "DEMONION" > ${TEST_FILE2}
    echo "Dungeon Crusadearz" > ${TEST_FILE3}

    echo "5f2864b5833190b07b0b95228682ff5ec43a13a2a3f31514c57d5c92aa3fb2e7  /tmp/mimixbox/it/sha256sum/1.txt" > ${CHECK_SUM_FILE}
    echo "833d8136112b60552a0f83165a2ebffeac4b0c0249480d651ea58b9073ec925b  /tmp/mimixbox/it/sha256sum/2.txt" >> ${CHECK_SUM_FILE}
    echo "8e774f75a5a23c83e6f7d5e92863a2615e0335e06aec18d9c3ec1c5315d1a777  /tmp/mimixbox/it/sha256sum/3.txt" >> ${CHECK_SUM_FILE}
}

Cleanup() {
    rm -rf ${TEST_DIR}
}

TestMd5sumOneFile() {
    sha256sum ${TEST_FILE1}
}

TestMd5sumOneDirectory() {
    sha256sum ${TEST_DIR}
}

TestMd5sumNotExistFile() {
    sha256sum /not_exist_file
}

TestMd5sumThreeFiles() {
    sha256sum ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
}

TestMd5sumWithCheckOption() {
    sha256sum -c ${CHECK_SUM_FILE}
}

TestMd5sumDataFromPipe() {
    echo "test" | sha256sum 
}

TestMd5sumFileAndDataFromPipeAtSameTime() {
    echo "test" | sha256sum ${TEST_FILE1}
}
