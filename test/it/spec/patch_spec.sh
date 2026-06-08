Describe 'patch'
    Include editors/patch_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'applies a unified diff'
        When call TestPatchApply
        The output should equal 'one
2
three'
        The status should be success
    End
End
