Setup() {
    export TEST_DIR=/tmp/mimixbox/it/dos2unix
    export TEST_FILE1=${TEST_DIR}/1.txt
    export TEST_FILE2=${TEST_DIR}/2.txt
    export TEST_FILE3=${TEST_DIR}/3.txt

    mkdir -p ${TEST_DIR}
    echo "abc"  > ${TEST_FILE1}
    echo "def" >> ${TEST_FILE1}
    echo "ghi" >> ${TEST_FILE1}
    sed -i 's/$/\r/g' ${TEST_FILE1}

    cp ${TEST_FILE1} ${TEST_FILE2}
    cp ${TEST_FILE1} ${TEST_FILE3}
}

Cleanup() {
    export TEST_DIR=/tmp/mimixbox/it/dos2unix
    rm -rf ${TEST_DIR}
}

TestDos2unixCRLF() {
    export TEST_DIR=/tmp/mimixbox/it/dos2unix
    export TEST_FILE1=${TEST_DIR}/1.txt

    dos2unix ${TEST_FILE1}
    file ${TEST_FILE1}
}

TestDos2unixCRLFStatus() {
    export TEST_DIR=/tmp/mimixbox/it/dos2unix
    export TEST_FILE1=${TEST_DIR}/1.txt

    dos2unix ${TEST_FILE1}
}

TestDos2unixThreeFileAtSameTime() {
    export TEST_DIR=/tmp/mimixbox/it/dos2unix
    export TEST_FILE1=${TEST_DIR}/1.txt
    export TEST_FILE2=${TEST_DIR}/2.txt
    export TEST_FILE3=${TEST_DIR}/3.txt

    dos2unix ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
    file ${TEST_FILE1}
    file ${TEST_FILE2}
    file ${TEST_FILE3}
}

TestDos2unixThreeFileAtSameTime() {
    export TEST_DIR=/tmp/mimixbox/it/dos2unix
    export TEST_FILE1=${TEST_DIR}/1.txt
    export TEST_FILE2=${TEST_DIR}/2.txt
    export TEST_FILE3=${TEST_DIR}/3.txt

    dos2unix ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
    file ${TEST_FILE1}
    file ${TEST_FILE2}
    file ${TEST_FILE3}
}

TestDos2unixThreeFileAtSameTimeStatus() {
    export TEST_DIR=/tmp/mimixbox/it/dos2unix
    export TEST_FILE1=${TEST_DIR}/1.txt
    export TEST_FILE2=${TEST_DIR}/2.txt
    export TEST_FILE3=${TEST_DIR}/3.txt

    dos2unix ${TEST_FILE1} ${TEST_FILE2} ${TEST_FILE3}
}

TestDos2unixDir() {
    export TEST_DIR=/tmp/mimixbox/it/dos2unix
    dos2unix ${TEST_DIR}
}

TestDos2unixOneOfThreeFail() {
    export TEST_DIR=/tmp/mimixbox/it/dos2unix
    export TEST_FILE1=${TEST_DIR}/1.txt
    export TEST_FILE2=${TEST_DIR}/2.txt
    export TEST_FILE3=${TEST_DIR}/3.txt

    dos2unix ${TEST_FILE1} ${TEST_DIR} ${TEST_FILE3} 
}