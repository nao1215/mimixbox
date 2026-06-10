Describe 'umount'
    Include util-linux/umount_test.sh

    It 'fails for a target that is not mounted'
        When call TestUmountNotMounted
        The output should equal 'rc=1'
        The status should be success
    End
    It 'requires a target'
        When call TestUmountNoArg
        The output should equal 'rc=1'
        The status should be success
    End
End
