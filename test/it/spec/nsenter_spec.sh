Describe 'nsenter'
    Include util-linux/nsenter_test.sh

    It 'requires a target PID'
        When call TestNsenterNoTarget
        The output should equal 'rc=1'
        The status should be success
    End
    It 'requires a namespace flag'
        When call TestNsenterNoNs
        The output should equal 'rc=1'
        The status should be success
    End
End
