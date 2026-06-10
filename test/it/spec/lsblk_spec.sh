Describe 'lsblk'
    Include util-linux/lsblk_test.sh

    It 'prints the column header'
        When call TestLsblkHeader
        The status should be success
        The output should include 'NAME'
        The output should include 'SIZE'
        The output should include 'TYPE'
    End
    It 'runs and exits successfully'
        When call TestLsblkRuns
        The output should equal '0'
        The status should be success
    End
End
