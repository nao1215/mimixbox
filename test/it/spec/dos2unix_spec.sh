Describe 'dos2unix CRLF file'
    Include textutils/dos2unix_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() {
        r="${MIMIXBOX_IT_ROOT}"
        printf '%s\n%s\n' \
          "dos2unix: converting file $r/dos2unix/1.txt to Unix format..." \
          "$r/dos2unix/1.txt: ASCII text"
    }

    It 'change LF'
        When call TestDos2unixCRLF
        The output should equal "$(result)"
    End
End

Describe 'Check status after dos2unix CRLF file'
    Include textutils/dos2unix_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() {
        r="${MIMIXBOX_IT_ROOT}"
        printf '%s\n' \
          "dos2unix: converting file $r/dos2unix/1.txt to Unix format..."
    }

    It 'status success'
        When call TestDos2unixCRLFStatus
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'dos2unix three CRLF file'
    Include textutils/dos2unix_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() {
        r="${MIMIXBOX_IT_ROOT}"
        printf '%s\n%s\n%s\n%s\n%s\n%s\n' \
          "dos2unix: converting file $r/dos2unix/1.txt to Unix format..." \
          "dos2unix: converting file $r/dos2unix/2.txt to Unix format..." \
          "dos2unix: converting file $r/dos2unix/3.txt to Unix format..." \
          "$r/dos2unix/1.txt: ASCII text" \
          "$r/dos2unix/2.txt: ASCII text" \
          "$r/dos2unix/3.txt: ASCII text"
    }

    It 'change LF'
        When call TestDos2unixThreeFileAtSameTime
        The output should equal "$(result)"
    End
End

Describe 'Check status after dos2unix three CRLF file'
    Include textutils/dos2unix_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() {
        r="${MIMIXBOX_IT_ROOT}"
        printf '%s\n%s\n%s\n' \
          "dos2unix: converting file $r/dos2unix/1.txt to Unix format..." \
          "dos2unix: converting file $r/dos2unix/2.txt to Unix format..." \
          "dos2unix: converting file $r/dos2unix/3.txt to Unix format..."
    }

    It 'status success'
        When call TestDos2unixThreeFileAtSameTimeStatus
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'Dos2unix directory'
    Include textutils/dos2unix_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() {
        r="${MIMIXBOX_IT_ROOT}"
        printf '%s\n' \
          "dos2unix: skip $r/dos2unix: not regular file"
    }

    It 'print "no regular file error"'
        When call TestDos2unixDir
        The error should equal "$(result)"
        The status should be failure
    End
End

Describe 'Dos2unix two file and one directory'
    Include textutils/dos2unix_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    output_result() {
        r="${MIMIXBOX_IT_ROOT}"
        printf '%s\n%s\n' \
          "dos2unix: converting file $r/dos2unix/1.txt to Unix format..." \
          "dos2unix: converting file $r/dos2unix/3.txt to Unix format..."
    }

    error_result() {
        r="${MIMIXBOX_IT_ROOT}"
        printf '%s\n' \
          "dos2unix: skip $r/dos2unix: not regular file"
    }


    It 'status error'
        When call TestDos2unixOneOfThreeFail
        The output should equal "$(output_result)"
        The error should equal "$(error_result)"
        The status should be failure
    End
End
