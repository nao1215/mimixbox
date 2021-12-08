Describe 'Word Count without options'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "  6  16 126 /tmp/mimixbox/it/game.txt"'
        When call TestWcWithNoOption
        The output should equal '  6  16 126 /tmp/mimixbox/it/game.txt'
        The status should be success
    End
End

Describe 'Word Count with --lines options'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "6 /tmp/mimixbox/it/game.txt"'
        When call TestWcWithLinesOption
        The output should equal '6 /tmp/mimixbox/it/game.txt'
        The status should be success
    End
End

Describe 'Word Count with --bytes options'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "126 /tmp/mimixbox/it/game.txt"'
        When call TestWcWithBytesOption
        The output should equal '126 /tmp/mimixbox/it/game.txt'
        The status should be success
    End
End

Describe 'Word Count with --max-line-length options'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "35 /tmp/mimixbox/it/game.txt"'
        When call TestWcWithMaxLineLengthOption
        The output should equal '35 /tmp/mimixbox/it/game.txt'
        The status should be success
    End
End

Describe 'Word Count for empty file'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "0 0 0 /tmp/mimixbox/it/empty.txt"'
        When call TestWcReadingEmptyFile
        The output should equal '0 0 0 /tmp/mimixbox/it/empty.txt'
        The status should be success
    End
End

Describe 'Word Count for three file'
    Include textutils/wc_test.sh
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
        The status should be success
    End
End

Describe 'Word Count from pipe'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'count file from pipe'
        When call TestWcWithPipe
        The output should equal '      1       1      26 '
        The status should be success
    End
End

Describe 'Word Count only file, not count pipe data'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'read the file with the specified argument without reading the pipe data'
        When call TestWcWithPipeAndArgument
        The output should equal '  6  16 126 /tmp/mimixbox/it/game.txt'
        The status should be success
    End
End

Describe 'Try word count directory'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'can not read directory'
        When call TestWcNotFile
        The error should equal 'wc: /tmp/mimixbox/it: this path is directory'
        The output should equal '      0       0       0 /tmp/mimixbox/it'
        The status should be failure
    End
End

Describe 'Try word count directory and file same time.'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() { %text
      #|      0       0       0 /tmp/mimixbox/it
      #|      6      16     126 /tmp/mimixbox/it/game.txt
      #|      6      16     126 total
    }

    It 'count only file. Count data for directory is always 0'
        When call TestWcDirectoryAndFileSameTime
        The error should equal 'wc: /tmp/mimixbox/it: this path is directory'
        The output should equal  "$(result)"
        The status should be failure
    End
End

Describe 'Count line from pipe data'
    Include textutils/wc_test.sh

    It 'say 1'
        When call TestWcNoExistFileNameFromPipeWithLinesOption
        The output should equal '1 '
        The status should be success
    End
End