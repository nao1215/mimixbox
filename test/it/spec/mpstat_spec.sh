Describe 'mpstat'
    Include procps/mpstat_test.sh

    It 'prints the CPU column header'
        When call TestMpstatHeader
        The output should include '%usr'
        The output should include '%idle'
        The status should be success
    End
    It 'prints the aggregate all row'
        When call TestMpstatAll
        The output should equal '1'
        The status should be success
    End
End
