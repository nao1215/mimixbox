Describe 'chmod with an octal mode'
    Include shellutils/chmod_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'sets the permission bits'
        When call TestChmodOctal
        The output should equal '644'
        The status should be success
    End
End

Describe 'chmod with a symbolic mode'
    Include shellutils/chmod_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'adds owner execute to mode 600'
        When call TestChmodSymbolic
        The output should equal '700'
        The status should be success
    End
End

Describe 'chmod on a missing file'
    Include shellutils/chmod_test.sh

    It 'reports an error'
        When call TestChmodMissing
        The error should equal "chmod: cannot access '/no_such_file': No such file or directory"
        The status should be failure
    End
End
