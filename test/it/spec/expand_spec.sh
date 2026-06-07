Describe 'expand from pipe'
    Include textutils/expand_test.sh

    It 'converts tabs to spaces (default tab stop 8)'
        When call TestExpandPipe
        The output should equal 'a       b'
        The status should be success
    End
End

Describe 'expand with a custom tab stop'
    Include textutils/expand_test.sh

    It 'converts tabs to the given width'
        When call TestExpandTabStop
        The output should equal 'a   b'
        The status should be success
    End
End

Describe 'expand from file'
    Include textutils/expand_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'converts tabs in the file'
        When call TestExpandFile
        The output should equal 'a       b'
        The status should be success
    End
End

Describe 'expand with a non-existent file'
    Include textutils/expand_test.sh

    It 'reports an error'
        When call TestExpandNoExistFile
        The error should equal 'expand: /no_exist_file: no such file or directory'
        The status should be failure
    End
End
