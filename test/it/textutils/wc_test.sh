export TEST_DIR=/tmp/mimixbox/it
export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
export TEST_FILE_METAL=/tmp/mimixbox/it/metal.txt
export EMPTY_FILE=/tmp/mimixbox/it/empty.txt

Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    export TEST_FILE_METAL=/tmp/mimixbox/it/metal.txt
    export EMPTY_FILE=/tmp/mimixbox/it/empty.txt

    mkdir -p /tmp/mimixbox/it

    echo "NieR Replicant ver.1.22474487139..." > /tmp/mimixbox/it/game.txt
    echo "NieR:Automata" >>  /tmp/mimixbox/it/game.txt
    echo "The Legend of Zelda: Majora's Mask" >>  /tmp/mimixbox/it/game.txt
    echo "KICHIKUOU RANCE" >>  /tmp/mimixbox/it/game.txt
    echo "DARK SOULS" >> /tmp/mimixbox/it/game.txt
    echo "SHADOW HEARTS" >>  /tmp/mimixbox/it/game.txt

    echo "MEGADETH" > $/tmp/mimixbox/it/metal.txt
    echo "GALNERYUS" >> /tmp/mimixbox/it/metal.txt
    echo "SYSTEM OF A DOWN" >> /tmp/mimixbox/it/metal.txt
    
    touch /tmp/mimixbox/it/empty.txt
}

CleanUp() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    export TEST_FILE_METAL=/tmp/mimixbox/it/metal.txt
    export EMPTY_FILE=/tmp/mimixbox/it/empty.txt

    rm  /tmp/mimixbox/it/empty.txt /tmp/mimixbox/it/game.txt /tmp/mimixbox/it/empty.txt
}

TestWcWithNoOption() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    wc /tmp/mimixbox/it/game.txt

TestWcWithLinesOption() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    wc -l /tmp/mimixbox/it/game.txt
}

TestWcWithBytesOption() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    wc -c /tmp/mimixbox/it/game.txt
}

TestWcWithMaxLineLengthOption() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    wc -L /tmp/mimixbox/it/game.txt
}

TestWcReadingEmptyFile() {
    export EMPTY_FILE=/tmp/mimixbox/it/empty.txt
    wc /tmp/mimixbox/it/empty.txt
}

TestWcReadingThreeFile() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    export TEST_FILE_METAL=/tmp/mimixbox/it/metal.txt
    export EMPTY_FILE=/tmp/mimixbox/it/empty.txt
    wc /tmp/mimixbox/it/empty.txt /tmp/mimixbox/it/game.txt /tmp/mimixbox/it/empty.txt
}
