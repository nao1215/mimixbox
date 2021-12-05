export TEST_DIR=/tmp/mimixbox/it/md5sum
export TEST_FILE1=${TEST_DIR}/1.txt
export TEST_FILE2=${TEST_DIR}/2.txt
export TEST_FILE3=${TEST_DIR}/3.txt
export CHECK_SUM_FILE=${TEST_DIR}/checksum.txt

Setup() {
    mkdir -p ${TEST_DIR}
    echo "Dungeon of Regalias" > ${TEST_FILE1} 
    echo "DEMONION" > ${TEST_FILE2}
    echo "Dungeon Crusadearz" > ${TEST_FILE3}

    echo "d0d8ffef81b3c7160ac655d5939548c5  /tmp/mimixbox/it/md5sum/1.txt" > ${CHECK_SUM_FILE}
    echo "07e280ad4bd77b9321f0ce3386775019  /tmp/mimixbox/it/md5sum/2.txt" >> ${CHECK_SUM_FILE}
    echo "15e924f84517598e828f49dc85765bc5  /tmp/mimixbox/it/md5sum/3.txt" >> ${CHECK_SUM_FILE}
}

Cleanup() {
    rm -rf ${TEST_DIR}
}

TestMd5sumOneFile() {
    md5sum ${TEST_FILE1}
}

TestMd5sumOneDirectory() {
    md5sum ${TEST_DIR}
}

TestMd5sumNotExistFile() {
    md5sum /not_exist_file
}

TestMd5sumThreeFiles() {
    md5sum ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
}

TestMd5sumWithCheckOption() {
    md5sum -c ${CHECK_SUM_FILE}
}

TestMd5sumDataFromPipe() {
    echo "test" | md5sum 
}

TestMd5sumFileAndDataFromPipeAtSameTime() {
    echo "test" | md5sum ${TEST_FILE1}
}
