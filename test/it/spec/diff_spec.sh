Describe 'diff'
    Include editors/diff_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'reports a change in normal format'
        When call TestDiffNormal
        The line 1 of output should equal '2c2'
        The status should equal 1
    End
    It 'is silent and succeeds for identical files'
        When call TestDiffSame
        The output should equal ''
        The status should be success
    End
    It 'reports briefly with -q'
        When call TestDiffBrief
        The output should include 'differ'
        The status should equal 1
    End
End
