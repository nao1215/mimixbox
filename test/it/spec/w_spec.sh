Describe 'w'
    Include shellutils/w_test.sh

    It 'prints a summary header with the load averages'
        When call TestWHeader
        The status should be success
        The output should include 'load average:'
        The output should include 'up '
    End
    It 'prints the column header'
        When call TestWColumns
        The status should be success
        The output should include 'USER'
        The output should include 'LOGIN@'
    End
End
