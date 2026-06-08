Describe 'split by lines'
    Include textutils/split_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    result() { %text
        #|1
        #|2
    }
    It 'splits input into files of N lines'
        When call TestSplitLines
        The output should equal "$(result)"
        The status should be success
    End
End
