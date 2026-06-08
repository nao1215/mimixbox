Describe 'unshadow'
    Include securityutils/unshadow_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'merges the shadow hash into the passwd line'
        When call TestUnshadow
        The output should include 'alice:$6$abc$HASH:1000:1000'
        The status should be success
    End
End
