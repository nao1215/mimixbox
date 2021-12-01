Describe 'Word Count without options'
    Include it/textutils/wc_test.sh

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

    echo "MEGADETH" > /tmp/mimixbox/it/metal.txt
    echo "GALNERYUS" >> /tmp/mimixbox/it/metal.txt
    echo "SYSTEM OF A DOWN" >> /tmp/mimixbox/it/metal.txt
    
    touch /tmp/mimixbox/it/empty.txt
}

CleanUp() {
    export TEST_FILE_GAMENAME=/tmp/mimixbox/it/game.txt
    export TEST_FILE_METAL=/tmp/mimixbox/it/metal.txt
    export EMPTY_FILE=/tmp/mimixbox/it/empty.txt

    rm  /tmp/mimixbox/it/empty.txt /tmp/mimixbox/it/game.txt /tmp/mimixbox/it/metal.txt
}

    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "  6  16 126 /tmp/mimixbox/it/game.txt"'
        When call TestWcWithNoOption
        The output should equal '  6  16 126 /tmp/mimixbox/it/game.txt'
    End
End

Describe 'Word Count with --lines options'
    Include it/textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "6 /tmp/mimixbox/it/game.txt"'
        When call TestWcWithLinesOption
        The output should equal '6 /tmp/mimixbox/it/game.txt'
    End
End

Describe 'Word Count with --bytes options'
    Include it/textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "126 /tmp/mimixbox/it/game.txt"'
        When call TestWcWithBytesOption
        The output should equal '126 /tmp/mimixbox/it/game.txt'
    End
End

Describe 'Word Count with --max-line-length options'
    Include it/textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "35 /tmp/mimixbox/it/game.txt"'
        When call TestWcWithMaxLineLengthOption
        The output should equal '35 /tmp/mimixbox/it/game.txt'
    End
End

Describe 'Word Count for empty file'
    Include it/textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "0 0 0 /tmp/mimixbox/it/empty.txt"'
        When call TestWcReadingEmptyFile
        The output should equal '0 0 0 /tmp/mimixbox/it/empty.txt'
    End
End

Describe 'Word Count for two file'
    Include it/textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
      #|  0   0   0 /tmp/mimixbox/it/empty.txt
      #|  6  16 126 /tmp/mimixbox/it/game.txt
      #|  3   6  36 /tmp/mimixbox/it/metal.txt
      #|  9  22 162 total
    }

    It 'says "wc: three file results"'
        When call TestWcReadingThreeFile
        The output should equal "$(result)"
    End
End