TEST_DIR=/tmp/mimixbox/it
TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
TEST_FILE_METAL=/tmp/mimixbox/it/metal.txt
EMPTY_FILE=/tmp/mimixbox/it/empty.txt

Setup() {
    mkdir -p ${TEST_DIR}

    echo "NieR Replicant ver.1.22474487139..." > ${TEST_FILE_GAMENAME}
    echo "NieR:Automata" >>  ${TEST_FILE_GAMENAME}
    echo "The Legend of Zelda: Majora's Mask" >>  ${TEST_FILE_GAMENAME}
    echo "KICHIKUOU RANCE" >>  ${TEST_FILE_GAMENAME}
    echo "DARK SOULS" >>  ${TEST_FILE_GAMENAME}
    echo "SHADOW HEARTS" >>  ${TEST_FILE_GAMENAME}

    echo "MEGADETH" > ${TEST_FILE_METAL}
    echo "GALNERYUS" >> ${TEST_FILE_METAL}
    echo "SYSTEM OF A DOWN" >> ${TEST_FILE_METAL}
    
    touch ${EMPTY_FILE}
}

CleanUp() {
    rm  ${TEST_FILE_GAMENAME} ${EMPTY_FILE}
}

TestWcWithNoOption() {
    mimixbox wc ${TEST_FILE_GAMENAME}
}

TestWcWithLinesOption() {
    mimixbox wc -l ${TEST_FILE_GAMENAME}
}

TestWcWithBytesOption() {
    mimixbox wc -c ${TEST_FILE_GAMENAME}
}

TestWcWithMaxLineLengthOption() {
    mimixbox wc -L ${TEST_FILE_GAMENAME}
}

TestWcReadingEmptyFile() {
    mimixbox wc ${EMPTY_FILE}
}

TestWcReadingThreeFile() {
    mimixbox wc ${EMPTY_FILE} ${TEST_FILE_GAMENAME} ${TEST_FILE_METAL}
}
