Describe 'Word Count without options'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "  6  16 126 ${MIMIXBOX_IT_ROOT}/game.txt"'
        When call TestWcWithNoOption
        The output should equal "  6  16 126 ${MIMIXBOX_IT_ROOT}/game.txt"
        The status should be success
    End
End

Describe 'Word Count with --lines options'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "6 ${MIMIXBOX_IT_ROOT}/game.txt"'
        When call TestWcWithLinesOption
        The output should equal "6 ${MIMIXBOX_IT_ROOT}/game.txt"
        The status should be success
    End
End

Describe 'Word Count with --bytes options'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "126 ${MIMIXBOX_IT_ROOT}/game.txt"'
        When call TestWcWithBytesOption
        The output should equal "126 ${MIMIXBOX_IT_ROOT}/game.txt"
        The status should be success
    End
End

Describe 'Word Count with --max-line-length options'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "35 ${MIMIXBOX_IT_ROOT}/game.txt"'
        When call TestWcWithMaxLineLengthOption
        The output should equal "35 ${MIMIXBOX_IT_ROOT}/game.txt"
        The status should be success
    End
End

Describe 'Word Count for empty file'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "0 0 0 ${MIMIXBOX_IT_ROOT}/empty.txt"'
        When call TestWcReadingEmptyFile
        The output should equal "0 0 0 ${MIMIXBOX_IT_ROOT}/empty.txt"
        The status should be success
    End
End

Describe 'Word Count for three file'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() {
        r="${MIMIXBOX_IT_ROOT}"
        printf '%s\n%s\n%s\n%s\n' \
          "  0   0   0 $r/empty.txt" \
          "  6  16 126 $r/game.txt" \
          "  3   6  36 $r/metal.txt" \
          "  9  22 162 total"
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

    # The byte count depends on the length of the per-run path that is echoed
    # into wc (path + trailing newline), so compute it instead of hardcoding it.
    pipe_result() {
        bytes=$(( ${#MIMIXBOX_IT_ROOT} + ${#_pipe_suffix} + 1 ))
        printf '%7d %7d %7d' 1 1 "$bytes"
    }

    It 'count file from pipe'
        _pipe_suffix='/game.txt'
        When call TestWcWithPipe
        The output should equal "$(pipe_result)"
        The status should be success
    End
End

Describe 'Word Count only file, not count pipe data'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'read the file with the specified argument without reading the pipe data'
        When call TestWcWithPipeAndArgument
        The output should equal "  6  16 126 ${MIMIXBOX_IT_ROOT}/game.txt"
        The status should be success
    End
End

Describe 'Try word count directory'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'can not read directory'
        When call TestWcNotFile
        The error should equal "wc: ${MIMIXBOX_IT_ROOT}: is a directory"
        The output should equal "      0       0       0 ${MIMIXBOX_IT_ROOT}"
        The status should be failure
    End
End

Describe 'Try word count directory and file same time.'
    Include textutils/wc_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    result() {
        r="${MIMIXBOX_IT_ROOT}"
        printf '%s\n%s\n%s\n' \
          "      0       0       0 $r" \
          "      6      16     126 $r/game.txt" \
          "      6      16     126 total"
    }

    It 'count only file. Count data for directory is always 0'
        When call TestWcDirectoryAndFileSameTime
        The error should equal "wc: ${MIMIXBOX_IT_ROOT}: is a directory"
        The output should equal  "$(result)"
        The status should be failure
    End
End

Describe 'Count line from pipe data'
    Include textutils/wc_test.sh

    It 'say 1'
        When call TestWcNoExistFileNameFromPipeWithLinesOption
        The output should equal '1'
        The status should be success
    End
End