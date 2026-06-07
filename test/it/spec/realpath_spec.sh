Describe 'realpath resolves an existing file'
    Include shellutils/realpath_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'prints the absolute path'
        When call TestRealpathExisting
        The output should equal '/tmp/mimixbox/it/realpath/file.txt'
        The status should be success
    End
End

Describe 'realpath -m on a missing path'
    Include shellutils/realpath_test.sh

    It 'prints the cleaned absolute path'
        When call TestRealpathMissing
        The output should equal '/tmp/mimixbox/it/realpath/does/not/exist'
        The status should be success
    End
End

Describe 'realpath with no operand'
    Include shellutils/realpath_test.sh

    It 'reports an error'
        When call TestRealpathNoOperand
        The error should equal 'realpath: missing operand'
        The status should be failure
    End
End
