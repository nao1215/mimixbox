Describe 'lsof'
    Include procps/lsof_test.sh

    It 'lists the working directory of a process'
        When call TestLsofSelf
        The status should be success
        The output should equal '1'
    End
    It 'prints the column header'
        When call TestLsofHeader
        The status should be success
        The output should include 'COMMAND'
        The output should include 'NAME'
    End
End
