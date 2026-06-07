Describe 'ln creates a hard link'
    Include fileutils/ln_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'links to the same content'
        When call TestLnHard
        The output should equal 'content'
        The status should be success
    End
End

Describe 'ln -s creates a symbolic link'
    Include fileutils/ln_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'creates a symlink'
        When call TestLnSymbolic
        The output should equal 'is symlink'
        The status should be success
    End
End

Describe 'ln with no operand'
    Include fileutils/ln_test.sh

    It 'reports an error'
        When call TestLnNoOperand
        The error should equal 'ln: missing file operand'
        The status should be failure
    End
End
