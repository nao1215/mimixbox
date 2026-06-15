Describe 'rmdir removes an empty directory'
    Include fileutils/rmdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'removes it'
        When call TestRmdirEmpty
        The output should equal 'removed'
        The status should be success
    End
End

Describe 'rmdir on a non-empty directory'
    Include fileutils/rmdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'fails with directory not empty'
        When call TestRmdirNonEmpty
        The error should equal "rmdir: failed to remove '${MIMIXBOX_IT_ROOT}/rmdir/full': Directory not empty"
        The status should be failure
    End
End

Describe 'rmdir -p removes nested empty directories'
    Include fileutils/rmdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'removes the whole chain'
        When call TestRmdirParents
        The output should equal 'removed'
        The status should be success
    End
End

Describe 'rmdir with no operand'
    Include fileutils/rmdir_test.sh

    It 'reports an error'
        When call TestRmdirMissingOperand
        The error should equal 'rmdir: missing operand'
        The status should be failure
    End
End
