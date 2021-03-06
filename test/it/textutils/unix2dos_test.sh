export TEST_DIR=/tmp/mimixbox/it/unix2dos
export TEST_FILE1=${TEST_DIR}/1.txt
export TEST_FILE2=${TEST_DIR}/2.txt
export TEST_FILE3=${TEST_DIR}/3.txt


Setup() {
    mkdir -p ${TEST_DIR}
    echo "abc"  > ${TEST_FILE1}
    echo "def" >> ${TEST_FILE1}
    echo "ghi" >> ${TEST_FILE1}

    cp ${TEST_FILE1} ${TEST_FILE2}
    cp ${TEST_FILE1} ${TEST_FILE3}
}

Cleanup() {
    rm -rf ${TEST_DIR}
}

TestUnix2dosCRLF() {
    /usr/local/bin/unix2dos ${TEST_FILE1}
    file ${TEST_FILE1}
}

TestUnix2dosCRLFStatus() {
    /usr/local/bin/unix2dos ${TEST_FILE1}
}

TestUnix2dosThreeFileAtSameTime() {
    /usr/local/bin/unix2dos ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
    file ${TEST_FILE1}
    file ${TEST_FILE2}
    file ${TEST_FILE3}
}

TestUnix2dosThreeFileAtSameTime() {
    /usr/local/bin/unix2dos ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
    file ${TEST_FILE1}
    file ${TEST_FILE2}
    file ${TEST_FILE3}
}

TestUnix2dosThreeFileAtSameTimeStatus() {
    /usr/local/bin/unix2dos ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
}

TestUnix2dosDir() {
    /usr/local/bin/unix2dos ${TEST_DIR}
}

TestUnix2dosOneOfThreeFail() {
    /usr/local/bin/unix2dos ${TEST_FILE1}  ${TEST_DIR} ${TEST_FILE3} 
}