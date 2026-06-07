Describe 'df prints a header'
    Include shellutils/df_test.sh

    It 'shows the column header'
        When call TestDfHeader
        The output should include 'Filesystem'
        The status should be success
    End
End

Describe 'df succeeds for the current directory'
    Include shellutils/df_test.sh

    It 'exits zero'
        When call TestDfStatus
        The output should equal 'rc=0'
        The status should be success
    End
End
