Describe 'ipcrm'
    Include util-linux/ipcrm_test.sh

    It 'fails when nothing is requested'
        When call TestIpcrmNothing
        The output should equal 'rc=1'
        The status should be success
    End
    It 'fails to remove a non-existent id'
        When call TestIpcrmBadId
        The output should equal 'rc=1'
        The status should be success
    End
End
