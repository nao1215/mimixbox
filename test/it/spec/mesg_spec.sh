Describe 'mesg'
    Include util-linux/mesg_test.sh

    It 'reports an error when stdin is not a terminal'
        When call TestMesgNotTty
        The output should equal 'mesg: cannot get terminal name
rc=2'
        The status should be success
    End
End
