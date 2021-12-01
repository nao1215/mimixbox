Describe 'Word Count without options'
    Include it/textutils/wc_test.sh

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