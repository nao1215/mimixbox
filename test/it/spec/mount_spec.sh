Describe 'mount'
    Include util-linux/mount_test.sh

    It 'lists the root filesystem'
        When call TestMountListsRoot
        The status should be success
        The output should not equal '0'
    End
    It 'refuses to perform a mount'
        When call TestMountRejectsMount
        The output should equal 'rc=1'
        The status should be success
    End
End
