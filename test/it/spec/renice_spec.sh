Describe 'renice'
    Include util-linux/renice_test.sh

    It 'reports the priority change'
        When call TestRenice
        The output should equal '1'
        The status should be success
    End
    It 'rejects a non-numeric PID'
        When call TestReniceInvalid
        The output should equal 'rc=1'
        The status should be success
    End
End
