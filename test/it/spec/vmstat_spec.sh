Describe 'vmstat'
    Include procps/vmstat_test.sh

    It 'prints the column header'
        When call TestVmstatHeader
        The output should include 'swpd'
        The output should include 'free'
        The status should be success
    End
    It 'prints a numeric data row'
        When call TestVmstatData
        The output should equal '1'
        The status should be success
    End
End
