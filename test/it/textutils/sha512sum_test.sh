export TEST_DIR=/tmp/mimixbox/it/sha512sum
export TEST_FILE1=${TEST_DIR}/1.txt
export TEST_FILE2=${TEST_DIR}/2.txt
export TEST_FILE3=${TEST_DIR}/3.txt
export CHECK_SUM_FILE=${TEST_DIR}/checksum.txt

Setup() {
    mkdir -p ${TEST_DIR}
    echo "Dungeon of Regalias" > ${TEST_FILE1} 
    echo "DEMONION" > ${TEST_FILE2}
    echo "Dungeon Crusadearz" > ${TEST_FILE3}

    echo "05eec7dcf412f63d5a291d019f6b3d62d4f8f5236592815ed171f7d6d0a7969f65a589a092740bd04a2f181d7d5a27ff36808e04a69bd84a854aad0a01da3612  /tmp/mimixbox/it/sha512sum/1.txt" > ${CHECK_SUM_FILE}
    echo "cb2389a103184f607973b1acd073dc15310c8172b03f340a52bdc3843621cf9fbc6263c7dbbd786ceb0244f5147a83aa32ce09a485f544093b7fc5c7533e564f  /tmp/mimixbox/it/sha512sum/2.txt" >> ${CHECK_SUM_FILE}
    echo "3dafa5f1ec7f09cbe551dc0d4bdb153dedb81104b7e930b7c20733965f7ebb86ee2abea64b6bfa1c54045032865044a3feca5dcc89c28def410b2954094a1890  /tmp/mimixbox/it/sha512sum/3.txt" >> ${CHECK_SUM_FILE}
}

Cleanup() {
    rm -rf ${TEST_DIR}
}

TestMd5sumOneFile() {
    sha512sum ${TEST_FILE1}
}

TestMd5sumOneDirectory() {
    sha512sum ${TEST_DIR}
}

TestMd5sumNotExistFile() {
    sha512sum /not_exist_file
}

TestMd5sumThreeFiles() {
    sha512sum ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
}

TestMd5sumWithCheckOption() {
    sha512sum -c ${CHECK_SUM_FILE}
}

TestMd5sumDataFromPipe() {
    echo "test" | sha512sum 
}

TestMd5sumFileAndDataFromPipeAtSameTime() {
    echo "test" | sha512sum ${TEST_FILE1}
}
