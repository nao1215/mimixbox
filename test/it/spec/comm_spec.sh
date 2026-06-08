Describe 'comm common lines'
    Include textutils/comm_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'prints lines common to both files'
        When call TestCommBoth
        The output should equal 'banana'
        The status should be success
    End
End
