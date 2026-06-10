Describe 'rdate'
    Include util-linux/rdate_test.sh

    It 'fails when no host is given'
        When call TestRdateNoHost
        The output should equal 'rc=1'
        The status should be success
    End
    It 'fails when the host has no time service'
        When call TestRdateUnreachable
        The output should equal 'rc=1'
        The status should be success
    End
End
