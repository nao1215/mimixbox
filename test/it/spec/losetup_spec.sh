Describe 'losetup'
    Include util-linux/losetup_test.sh

    It 'lists active loop devices cleanly'
        When call TestLosetupAll
        The output should equal 'rc=0'
        The status should be success
    End
    It 'refuses to associate a loop device'
        When call TestLosetupSetup
        The output should equal 'rc=1'
        The status should be success
    End
End
