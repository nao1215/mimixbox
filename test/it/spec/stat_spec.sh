Describe 'stat'
    Include fileutils/stat_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'prints the size with a custom format'
        When call TestStatSize
        The output should equal '5'
        The status should be success
    End
End
