Describe 'pmap'
    Include procps/pmap_test.sh

    It 'prints a total line for a process map'
        When call TestPmapTotal
        The output should equal '1'
        The status should be success
    End
    It 'rejects an invalid PID'
        When call TestPmapInvalid
        The output should equal 'rc=1'
        The status should be success
    End
End
