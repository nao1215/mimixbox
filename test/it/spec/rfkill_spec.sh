Describe 'rfkill'
    Include console-tools/rfkill_test.sh

    It 'lists devices cleanly'
        When call TestRfkillList
        The output should equal 'rc=0'
        The status should be success
    End
    It 'rejects an unknown command'
        When call TestRfkillUnknown
        The output should equal 'rc=1'
        The status should be success
    End
End
