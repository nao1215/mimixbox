Describe 'Touch one file.'
    Include fileutils/touch_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'make touch.txt'
        When call TestTouchOneFile
        The output should equal "${MIMIXBOX_IT_ROOT}/touch/touch.txt"
    End
End

Describe 'Check status after making one file.'
    Include fileutils/touch_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'status success'
        When call TestTouchOneFileStatus
        The status should be success
    End
End

Describe 'Touch three file.'
    Include fileutils/touch_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() {
        r="${MIMIXBOX_IT_ROOT}"
        printf '%s\n%s\n%s\n' \
          "$r/touch/1.txt" \
          "$r/touch/2.txt" \
          "$r/touch/3.txt"
    }

    It 'make 1.txt 2.txt 3.txt'
        When call TestTouchThreeFileAtSameTime
        The output should equal "$(result)"
    End
End

Describe 'Check status after making three file.'
    Include fileutils/touch_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'status success'
        When call TestTouchThreeFileAtSameTimeStatus
        The status should be success
    End
End

Describe 'Touch three file and not make one file'
    Include fileutils/touch_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() {
        r="${MIMIXBOX_IT_ROOT}"
        printf '%s\n%s\n' \
          "$r/touch/1.txt" \
          "$r/touch/3.txt"
    }

    It 'make 1.txt 2.txt 3.txt'
        When call TestTouchThreeFileAndNotMakeOneFile
        The output should equal "$(result)"
        The error should equal "touch: open /touch/2.txt: no such file or directory"
    End
End

Describe 'Check sttus touch three file and not make one file'
    Include fileutils/touch_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'make 1.txt 2.txt 3.txt'
        When call TestTouchThreeFileAndNotMakeOneFileStatus
        The error should equal "touch: open /touch/2.txt: no such file or directory"
        The status should be failure
    End
End
