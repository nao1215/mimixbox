export TEST_DIR=/tmp/mimixbox/it
export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
export TEST_FILE_METAL=/tmp/mimixbox/it/metal.txt
export EMPTY_FILE=/tmp/mimixbox/it/empty.txt

Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    export TEST_FILE_METAL=/tmp/mimixbox/it/metal.txt
    export EMPTY_FILE=/tmp/mimixbox/it/empty.txt

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
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    export TEST_FILE_METAL=/tmp/mimixbox/it/metal.txt
    export EMPTY_FILE=/tmp/mimixbox/it/empty.txt

    rm  ${TEST_FILE_GAMENAME} ${EMPTY_FILE} ${TEST_FILE_METAL}
}

TestWcWithNoOption() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    echo ${TEST_FILE_GAMENAME}
    cat ${TEST_FILE_GAMENAME}
    wc ${TEST_FILE_GAMENAME}
}

TestWcWithLinesOption() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    wc -l ${TEST_FILE_GAMENAME}
}

TestWcWithBytesOption() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    wc -c ${TEST_FILE_GAMENAME}
}

TestWcWithMaxLineLengthOption() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    wc -L ${TEST_FILE_GAMENAME}
}

TestWcReadingEmptyFile() {
    export EMPTY_FILE=/tmp/mimixbox/it/empty.txt
    wc ${EMPTY_FILE}
}

TestWcReadingThreeFile() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    export TEST_FILE_METAL=/tmp/mimixbox/it/metal.txt
    export EMPTY_FILE=/tmp/mimixbox/it/empty.txt
    wc ${EMPTY_FILE} ${TEST_FILE_GAMENAME} ${TEST_FILE_METAL}
}