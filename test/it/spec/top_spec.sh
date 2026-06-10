Describe 'top'
    Include procps/top_test.sh

    It 'prints the top summary line'
        When call TestTopHeader
        The output should equal '1'
        The status should be success
    End
    It 'prints the tasks line'
        When call TestTopTasks
        The output should equal '1'
        The status should be success
    End
End
